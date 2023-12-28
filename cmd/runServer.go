/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"log"

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
	Short: "A brief description of your command",
	Long: `A longer description that spans multiple lines and likely contains examples
and usage of using your command. For example:

Cobra is a CLI library for Go that empowers applications.
This application is a tool to generate the needed files
to quickly create a Cobra application.`,
	Run: func(cmd *cobra.Command, args []string) {
		RunServer()
	},
}

func init() {
	rootCmd.AddCommand(runServerCmd)

	// Here you will define your flags and configuration settings.

	// Cobra supports Persistent Flags which will work for this command
	// and all subcommands, e.g.:
	// runServerCmd.PersistentFlags().String("foo", "", "A help for foo")

	// Cobra supports local flags which will only run when this command
	// is called directly, e.g.:
	// runServerCmd.Flags().BoolP("toggle", "t", false, "Help message for toggle")
}

func RunServer() {
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
	r := gin.Default()

	ctx := context.Background()
	// connecting to dependencies
	jwtHandler := jwt.NewGenerator(c.JWT)

	smtpHandler := smtp.NewSMTPHandler(c.SMTP)

	aesHandler, err := aes.NewAesHandler([]byte(c.AESKey))
	if err != nil {
		log.Fatal("error on creating aes handler", err)
	}

	otpStorage := otp.NewOTPStorage()

	dbConn, err := postgres.NewConnectionPool(ctx, c.Postgres)
	if err != nil {
		log.Fatal("error on connecting to db", err)
	}

	minioClient, err := minio.NewMinioHandler(ctx, c.MinIO)
	if err != nil {
		log.Fatal("error on getting new minio client", err)
	}

	// make repo
	authRepo, err := postgres.NewAuthRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating auth repo: ", err)
	}

	problemsMetadataRepo, err := postgres.NewProblemsMetadataRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating problems metadata repo: ", err)
	}

	problemsDescriptionRepo, err := mongodb.NewProblemDescriptionRepo(c.Mongo)
	if err != nil {
		log.Fatal("error on creating problem description repo: ", err)
	}

	submissionsRepo, err := postgres.NewSubmissionRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating submission metadata repo: ", err)
	}

	testcaseRepo, err := postgres.NewTestCaseRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating testcase repo: ", err)
	}

	judgeRepo, err := mongodb.NewJudgeRepo(c.Mongo)
	if err != nil {
		log.Fatal("error on creating judge repo")
	}

	contestRepo, err := postgres.NewContestsMetadataRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating contest repo", err)
	}

	contestsProblemsRepo, err := postgres.NewContestsProblemsMetadataRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating contest problems repo: ", err)
	}

	// initiating module handlers
	judgeHandler, err := judge.NewJudge(c.Judge, submissionsRepo, minioClient, testcaseRepo, judgeRepo)
	if err != nil {
		log.Fatal("error on creating judge handler", err)
	}
	authHandler := auth.NewAuthHandler(authRepo, jwtHandler, smtpHandler, c, aesHandler, otpStorage)
	problemsHandler := problems.NewProblemsHandler(problemsMetadataRepo, problemsDescriptionRepo, testcaseRepo)
	submissionsHandler := submissions.NewSubmissionsHandler(submissionsRepo, minioClient, judgeHandler)
	contestHandler := contests.NewContestsHandler(contestRepo, contestsProblemsRepo, problemsMetadataRepo)

	// starting http server
	api.AddRoutes(r, authHandler, problemsHandler, submissionsHandler, contestHandler)

	addr := fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
	pkg.Log.Info("Running on address: ", addr)
	if err := r.Run(addr); err != nil {
		panic(err)

	}
}
