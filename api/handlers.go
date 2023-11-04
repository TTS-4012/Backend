package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type handlers struct {
	authHandler auth.AuthHandler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler) {
	h := handlers{authHandler}

	v1 := r.Group("/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register_user", h.registerUser)
			authGroup.GET("/verify/:token", h.verifyEmail)
			authGroup.POST("/login_user", h.loginUser)
			authGroup.POST("/renew_token", h.renewToken)
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
	}

	resp, status, err := h.authHandler.RegisterUser(c, reqData)
	if err != nil {
		pkg.Log.Error("Failed to register user", err)
	}
	c.JSON(status, resp)
}

func (h *handlers) verifyEmail(c *gin.Context) {
	token := c.Param("token")
	status := h.authHandler.VerifyEmail(c, token)
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

	resp, status, err := h.authHandler.LoginUser(c, reqData)
	if err != nil {
		logger.Error("error on handling login", err)
	}
	c.JSON(status, resp)
}

func (h *handlers) renewToken(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "renewToken")

	oldRefreshToken, exists := c.Request.Header["Authorization"]
	if !exists {
		logger.Error("no refresh token provided")
		c.Status(http.StatusUnauthorized)
		return
	}
	if len(oldRefreshToken) != 1 {
		logger.Warning("multiple authorization values!", oldRefreshToken)
		c.Status(http.StatusBadRequest)
		return
	}

	resp, status, err := h.authHandler.RenewToken(c, oldRefreshToken[0])
	if err != nil {
		logger.Error("error on handling login", err)
	}
	c.JSON(status, resp)
}
