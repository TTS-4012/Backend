package api

import (
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/internal/oc/problems"
	"ocontest/internal/oc/submissions"

	"github.com/gin-gonic/gin"
)

type handlers struct {
	authHandler        auth.AuthHandler
	problemsHandler    problems.ProblemsHandler
	submissionsHandler submissions.Handler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler, problemHandler problems.ProblemsHandler, submissionsHandler submissions.Handler) {
	h := handlers{
		authHandler:        authHandler,
		problemsHandler:    problemHandler,
		submissionsHandler: submissionsHandler,
	}

	r.Use(h.corsHandler)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.OPTIONS("/*cors", func(c *gin.Context) {
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
			problemGroup.POST("", h.CreateProblem)
			problemGroup.GET("/:id", h.GetProblem)
			problemGroup.GET("", h.ListProblems)
			problemGroup.PUT("/:id", h.UpdateProblem)
			problemGroup.DELETE("/:id", h.DeleteProblem)
		}
		submissionGroup := v1.Group("/submissions", h.AuthMiddleware())
		{
			submissionGroup.GET("/:id", h.GetSubmission)
			submissionGroup.POST("/", h.Submit)
		}
	}
}
