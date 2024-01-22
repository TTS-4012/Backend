/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
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
	"github.com/ocontest/backend/internal/db/postgres"
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

	pgConn, err := postgres.NewConnectionPool(ctx, c.Postgres)
	if err != nil {
		log.Fatal("error on connecting to postgres", err)
	}

	mongoConn, err := mongodb.NewConn(ctx, c.Mongo)
	if err != nil {
		log.Fatal("error on connecting to mongo", err)
	}

	minioClient, err := minio.NewMinioHandler(ctx, c.MinIO)
	if err != nil {
		log.Fatal("error on getting new minio client", err)
	}

	// make repo
	authRepo, err := postgres.NewAuthRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating auth repo: ", err)
	}

	problemsMetadataRepo, err := postgres.NewProblemsMetadataRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating problems metadata repo: ", err)
	}

	problemsDescriptionRepo, err := mongodb.NewProblemDescriptionRepo(c.Mongo)
	if err != nil {
		log.Fatal("error on creating problem description repo: ", err)
	}

	submissionsRepo, err := postgres.NewSubmissionRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating submission metadata repo: ", err)
	}

	testcaseRepo, err := postgres.NewTestCaseRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating testcase repo: ", err)
	}

	judgeRepo, err := mongodb.NewJudgeRepo(mongoConn, c.Mongo.Database)
	if err != nil {
		log.Fatal("error on creating judge repo")
	}

	contestRepo, err := postgres.NewContestsMetadataRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating contest repo", err)
	}

	contestsProblemsRepo, err := postgres.NewContestsProblemsMetadataRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating contest problems repo: ", err)
	}

	contestsUsersRepo, err := postgres.NewContestsUsersRepo(ctx, pgConn)
	if err != nil {
		log.Fatal("error on creating contest users repo: ", err)
	}

	// initiating module handlers
	judgeHandler, err := judge.NewJudge(c.Judge, submissionsRepo, minioClient, testcaseRepo, contestsUsersRepo, judgeRepo)
	if err != nil {
		log.Fatal("error on creating judge handler", err)
	}
	authHandler := auth.NewAuthHandler(authRepo, jwtHandler, smtpHandler, c, aesHandler, otpHandler)
	problemsHandler := problems.NewProblemsHandler(problemsMetadataRepo, problemsDescriptionRepo, testcaseRepo)
	submissionsHandler := submissions.NewSubmissionsHandler(submissionsRepo, minioClient, judgeHandler)
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
				pgConn.Close()
				pkg.Log.Info("pg conn closed")
				return nil
			},
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
