package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/internal/oc/problems"
)

type handlers struct {
	authHandler     auth.AuthHandler
	problemsHandler problems.ProblemsHandler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler, problemHandler problems.ProblemsHandler) {
	h := handlers{
		authHandler:     authHandler,
		problemsHandler: problemHandler,
	}

	r.Use(h.Cors)
	r.GET("/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "pong",
		})
	})

	r.Use(h.corsHandler)
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
			problemGroup.POST("/", h.CreateProblem)
			problemGroup.GET("/:id", h.GetProblem)
			problemGroup.GET("/", h.ListProblems)
		}
	}
}
