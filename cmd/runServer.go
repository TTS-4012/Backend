/*
Copyright © 2023 NAME HERE <EMAIL ADDRESS>
*/
package cmd

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"log"
	"ocontest/api"
	"ocontest/internal/db/mongodb"
	"ocontest/internal/db/postgres"
	"ocontest/internal/judge"
	"ocontest/internal/jwt"
	"ocontest/internal/minio"
	"ocontest/internal/oc/auth"
	"ocontest/internal/oc/contests"
	"ocontest/internal/oc/problems"
	"ocontest/internal/oc/submissions"
	"ocontest/internal/otp"
	"ocontest/pkg"
	"ocontest/pkg/aes"
	"ocontest/pkg/configs"
	"ocontest/pkg/smtp"

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

	contestRepo, err := postgres.NewContestsMetadataRepo(ctx, dbConn)
	if err != nil {
		log.Fatal("error on creating contest repo")
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

	// initiating module handlers
	judgeHandler, err := judge.NewJudge(c.Judge, submissionsRepo, minioClient, testcaseRepo, judgeRepo)
	if err != nil {
		log.Fatal("error on creating judge handler", err)
	}
	authHandler := auth.NewAuthHandler(authRepo, jwtHandler, smtpHandler, c, aesHandler, otpStorage)
	problemsHandler := problems.NewProblemsHandler(problemsMetadataRepo, problemsDescriptionRepo)
	submissionsHandler := submissions.NewSubmissionsHandler(submissionsRepo, minioClient, judgeHandler)
	contestHandler := contests.NewContestsHandler(contestRepo)

	// starting http server
	api.AddRoutes(r, authHandler, problemsHandler, submissionsHandler, contestHandler)

	addr := fmt.Sprintf("%s:%s", c.Server.Host, c.Server.Port)
	pkg.Log.Info("Running on address: ", addr)
	if err := r.Run(addr); err != nil {
		panic(err)

	}
}
