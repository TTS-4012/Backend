package api

import (
	"github.com/ocontest/backend/internal/oc/auth"
	"github.com/ocontest/backend/internal/oc/contests"
	contestsProblems "github.com/ocontest/backend/internal/oc/contestsProblems"
	"github.com/ocontest/backend/internal/oc/problems"
	"github.com/ocontest/backend/internal/oc/submissions"
	"net/http"

	"github.com/gin-gonic/gin"
)

type handlers struct {
	authHandler             auth.AuthHandler
	problemsHandler         problems.ProblemsHandler
	contestsHandler         contests.ContestsHandler
	submissionsHandler      submissions.Handler
	contestsProblemsHandler contestsProblems.ContestsProblemsHandler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler, problemHandler problems.ProblemsHandler, submissionsHandler submissions.Handler,
	contestsHandler contests.ContestsHandler, contestsProblemsHandler contestsProblems.ContestsProblemsHandler) {
	h := handlers{
		authHandler:             authHandler,
		problemsHandler:         problemHandler,
		submissionsHandler:      submissionsHandler,
		contestsHandler:         contestsHandler,
		contestsProblemsHandler: contestsProblemsHandler,
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
		//contestGroup := v1.Group("/contests")
		contestGroup := v1.Group("/contests", h.AuthMiddleware())
		{
			contestGroup.POST("", h.CreateContest)
			contestGroup.GET("/:id", h.GetContest)
			contestGroup.GET("", h.ListContests)
			contestGroup.PUT("/:id", h.UpdateContest)
			contestGroup.DELETE("/:id", h.DeleteContest)
			contestGroup.POST("/add_problem", h.AddProblemContest)
		}
		submissionGroup := v1.Group("/submissions", h.AuthMiddleware())
		{
			submissionGroup.GET("/:id", h.GetSubmission)
			submissionGroup.POST("/", h.Submit)
			submissionGroup.GET("/:id/results", h.GetSubmissionResult)
		}
		v1.POST("/problems/:problem_id/submit", h.AuthMiddleware(), h.Submit)
		v1.GET("/problems/:id/submissions", h.AuthMiddleware(), h.ListSubmissions)
		v1.GET("/problems/:id/submissions/all", h.AuthMiddleware(), h.ListAllSubmissions)
	}
}
