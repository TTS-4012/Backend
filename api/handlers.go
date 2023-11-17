package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/internal/oc/problems"
	"ocontest/pkg"
	"ocontest/pkg/structs"
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

func (h *handlers) registerUser(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "registerUser")

	var reqData structs.RegisterUserRequest
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.authHandler.RegisterUser(c, reqData)
	c.JSON(status, resp)
}

func (h *handlers) verifyEmail(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "verifyEmail")

	var reqData structs.RequestWithOTPCreds
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	status := h.authHandler.VerifyEmail(c, reqData.UserID, reqData.OTP)
	c.Status(status)
}

func (h *handlers) loginUser(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "loginUser")
	var reqData structs.LoginUserRequest

	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Error("error on binding request data json")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.authHandler.LoginUser(c, reqData)
	c.JSON(status, resp)
}

func (h *handlers) renewToken(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "renewToken")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError,
		})
		return
	}
	tokenType, exists := c.Get(TokenTypeKey)
	if !exists {
		logger.Error("error on getting token type from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError,
		})
		return
	}

	full_refresh := c.GetHeader("full-refresh") == "true" // if this header is set to true, then the refresh token will be renewed too

	resp, status := h.authHandler.RenewToken(c, userID.(int64), tokenType.(string), full_refresh)
	c.JSON(status, resp)
}

func (h *handlers) requestOTPLogin(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "requestOTPLogin")

	var reqData structs.RequestWithOTPCreds
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	status := h.authHandler.RequestLoginWithOTP(c, reqData.UserID)
	c.Status(status)
}

func (h *handlers) checkOTPLogin(c *gin.Context) {

	logger := pkg.Log.WithField("handler", "checkOTPLogin")

	var reqData structs.RequestWithOTPCreds
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.authHandler.CheckLoginWithOTP(c, reqData.UserID, reqData.OTP)
	c.JSON(status, resp)
}

func (h *handlers) editUser(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "editUser")

	var reqData structs.RequestEditUser
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	status := h.authHandler.EditUser(c, reqData)
	c.Status(status)
}

func (h *handlers) CreateProblem(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "createProblem")

	var reqData structs.RequestCreateProblem
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	resp, status := h.problemsHandler.CreateProblem(c, reqData)
	c.JSON(status, resp)
}

func (h *handlers) GetProblem(c *gin.Context) {
	h.problemsHandler.GetProblem(c, 0)
}

func (h *handlers) ListProblems(c *gin.Context) {
	h.problemsHandler.ListProblem(c, structs.RequestListProblems{})
}
