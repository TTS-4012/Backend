/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/ocontest/backend/internal/db"
	"github.com/ocontest/backend/internal/db/repos"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/ocontest/backend/pkg/kvstorages"
	"golang.org/x/sync/errgroup"

	"github.com/gin-gonic/gin"
	"github.com/ocontest/backend/api"
	"github.com/ocontest/backend/internal/db/mongodb"
	"github.com/ocontest/backend/internal/judge"
	"github.com/ocontest/backend/internal/jwt"
	"github.com/ocontest/backend/internal/minio"
	"github.com/ocontest/backend/internal/oc/auth"
	"github.com/ocontest/backend/internal/oc/contests"
	"github.com/ocontest/backend/internal/oc/problems"
	"github.com/ocontest/backend/internal/oc/submissions"
	"github.com/ocontest/backend/internal/otp"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/aes"
	"github.com/ocontest/backend/pkg/configs"
	"github.com/ocontest/backend/pkg/smtp"

	"github.com/spf13/cobra"
)

// runServerCmd represents the runServer command
var runServerCmd = &cobra.Command{
	Use:   "runServer",
	Short: "will run server according to it's given config",
	Run: func(cmd *cobra.Command, args []string) {
		srv, shutdown := getServer()

		go func() {
			pkg.Log.Info("running server on ", srv.Addr)
			if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
				log.Fatalf("listen: %s\n", err)
			}
		}()

		quit := make(chan os.Signal)
		signal.Notify(quit, syscall.SIGINT, syscall.SIGQUIT, syscall.SIGTERM)
		<-quit

		log.Println("Shutdown Server ...")

		if err := shutdown(); err != nil {
			log.Fatal("Dependencies Shutdown:", err)
		}
		log.Println("Server exiting")
	},
}

func init() {
	rootCmd.AddCommand(runServerCmd)
}

func getServer() (*http.Server, func() error) {
	configs.InitConf()
	c := configs.Conf
	pkg.InitLog(c.Log)
	pkg.Log.Info("config and log modules initialized")

	fmt.Println(c.Judge)
	if c.Judge.EnableRunner {
		pkg.Log.Info("runner part will be running too!")
		go RunRunnerTaskHandler(c)
	}

	gin.SetMode(gin.ReleaseMode)

	ctx := context.Background()
	// connecting to dependencies
	jwtHandler := jwt.NewGenerator(c.JWT)

	smtpHandler := smtp.NewSMTPHandler(c.SMTP)

	aesHandler, err := aes.NewAesHandler([]byte(c.AESKey))
	if err != nil {
		log.Fatal("error on creating aes handler: ", err)
	}

	kvStore, err := kvstorages.NewKVStorage(c.KVStore)
	if err != nil {
		log.Fatal("coudn't initialize kvstore:  ", err)
	}

	otpHandler := otp.NewOTPHandler(kvStore)

	mongoConn, err := mongodb.NewConn(ctx, c.Mongo)
	if err != nil {
		log.Fatal("error on connecting to mongo", err)
	}

	minioClient, err := minio.NewMinioHandler(ctx, c.MinIO)
	if err != nil {
		log.Fatal("error on getting new minio client", err)
	}

	repoWrapper, err := db.NewRepoWrapper(ctx, c.SQLDB)
	if err != nil {
		log.Fatal("couldn't connect to db error: ", err)
	}
	// make repos
	var authRepo repos.UsersRepo
	err = repoWrapper(ctx, &authRepo)
	if err != nil {
		log.Fatal("error on creating auth repos: ", err)
	}

	var problemsMetadataRepo repos.ProblemsMetadataRepo
	err = repoWrapper(ctx, &problemsMetadataRepo)
	if err != nil {
		log.Fatal("error on creating problems metadata repos: ", err)
	}

	problemsDescriptionRepo, err := mongodb.NewProblemDescriptionRepo(c.Mongo)
	if err != nil {
		log.Fatal("error on creating problem description repos: ", err)
	}

	var submissionsRepo repos.SubmissionMetadataRepo
	err = repoWrapper(ctx, &submissionsRepo)
	if err != nil {
		log.Fatal("error on creating submission metadata repos: ", err)
	}

	var testcaseRepo repos.TestCaseRepo
	err = repoWrapper(ctx, &testcaseRepo)
	if err != nil {
		log.Fatal("error on creating testcase repos: ", err)
	}

	judgeRepo, err := mongodb.NewJudgeRepo(mongoConn, c.Mongo.Database)
	if err != nil {
		log.Fatal("error on creating judge repos")
	}

	var contestRepo repos.ContestsMetadataRepo
	err = repoWrapper(ctx, &contestRepo)
	if err != nil {
		log.Fatal("error on creating contest repos", err)
	}

	var contestsProblemsRepo repos.ContestsProblemsRepo
	err = repoWrapper(ctx, &contestsProblemsRepo)
	if err != nil {
		log.Fatal("error on creating contest problems repos: ", err)
	}

	var contestsUsersRepo repos.ContestsUsersRepo
	err = repoWrapper(ctx, &contestsUsersRepo)
	if err != nil {
		log.Fatal("error on creating contest users repos: ", err)
	}

	// initiating module handlers
	judgeHandler, err := judge.NewJudge(c.Judge, submissionsRepo, minioClient, testcaseRepo, contestsUsersRepo, judgeRepo)
	if err != nil {
		log.Fatal("error on creating judge handler", err)
	}
	authHandler := auth.NewAuthHandler(authRepo, jwtHandler, smtpHandler, c, aesHandler, otpHandler)
	problemsHandler := problems.NewProblemsHandler(problemsMetadataRepo, problemsDescriptionRepo, testcaseRepo)
	submissionsHandler := submissions.NewSubmissionsHandler(
		submissionsRepo,
		contestRepo, contestsProblemsRepo, contestsUsersRepo, minioClient, judgeHandler)
	contestHandler := contests.NewContestsHandler(
		contestRepo, contestsProblemsRepo, problemsMetadataRepo,
		submissionsRepo, authRepo, contestsUsersRepo, judgeHandler)

	r := gin.Default()
	// starting http server
	api.AddRoutes(r, authHandler, problemsHandler, submissionsHandler, contestHandler)

	srv := &http.Server{
		Addr:    fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port),
		Handler: r,
	}

	shutdown := func() error {

		pkg.Log.Info("shutting down")

		ctx := context.Background()
		ctx, cancel := context.WithTimeout(ctx, c.Server.GracefulShutdownPeriod)
		defer cancel()

		if err := srv.Shutdown(ctx); err != nil {
			return err
		}

		tasks := []func() error{
			func() error {
				err := kvStore.Close()
				pkg.Log.WithError(err).Info("kv store closed")
				return err
			},
			func() error {
				err := mongoConn.Disconnect(ctx)
				pkg.Log.WithError(err).Info("mongo conn closed")
				return err
			},
		}

		errGroup := &errgroup.Group{}

		for _, t := range tasks {
			errGroup.Go(t)
		}

		return errGroup.Wait()
	}

	return srv, shutdown
}
