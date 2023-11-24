package api

import (
	"net/http"
	"ocontest/internal/minio"
	"ocontest/internal/oc/auth"
	"ocontest/internal/oc/problems"

	"github.com/gin-gonic/gin"
)

type handlers struct {
	authHandler        auth.AuthHandler
	problemsHandler    problems.ProblemsHandler
	submissionsHandler minio.SubmissionsHandler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler, problemHandler problems.ProblemsHandler, submissionsHandler minio.SubmissionsHandler) {
	h := handlers{
		authHandler:        authHandler,
		problemsHandler:    problemHandler,
		submissionsHandler: submissionsHandler,
	}

	r.Use(h.Cors)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.OPTIONS("/*cors", func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, full-refresh")
		c.Status(http.StatusOK)
	})
	v1 := r.Group("/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register", h.registerUser)
			authGroup.POST("/verify", h.verifyEmail)
			authGroup.POST("/otp", h.GetOTPForLogin)
			authGroup.POST("/login", h.loginUser)
			authGroup.POST("/renew_token", h.AuthMiddleware(), h.renewToken)
			authGroup.POST("/edit_user", h.AuthMiddleware(), h.editUser)
		}
		problemGroup := v1.Group("/problems", h.AuthMiddleware())
		{
			problemGroup.POST("/", h.CreateProblem)
			problemGroup.GET("/:id", h.GetProblem)
			problemGroup.GET("/", h.ListProblems)
		}
		submissionGroup := v1.Group("/submission", h.AuthMiddleware())
		{
			submissionGroup.GET("/:id", h.DownloadSubmission)
			submissionGroup.POST("/", h.UploadSubmission)
		}
	}
}
