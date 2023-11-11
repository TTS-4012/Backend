package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/pkg"
	"ocontest/pkg/structs"
	"strings"
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
			authGroup.POST("/register", h.registerUser)
			authGroup.POST("/verify", h.verifyEmail)
			authGroup.POST("/login", h.loginUser)
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

	resp, status := h.authHandler.RegisterUser(c, reqData)
	c.JSON(status, resp)
}

func (h *handlers) verifyEmail(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "verifyEmail")

	var reqData structs.RequestVerifyUser
	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Warn("Failed to read request body", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
	}

	status, err := h.authHandler.VerifyEmail(c, reqData.UserID, reqData.OTP)
	if err != nil {
		logger.Error("failed to verify ")
	}
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

	authHeader, exists := c.Request.Header["Authorization"]
	if !exists {
		logger.Error("no refresh token provided")
		c.JSON(http.StatusUnauthorized, gin.H{
			"Ok":      false,
			"Message": "no refresh token provided",
		})
		return
	}
	if len(authHeader) != 1 {
		logger.Warning("multiple authorization values!", authHeader)
		c.JSON(http.StatusBadRequest, gin.H{
			"Ok":      false,
			"Message": "multiple authorization values!",
		})
		return
	}

	oldRefreshToken := strings.TrimSpace(strings.Replace(authHeader[0], "Bearer", "", 1))
	resp, status := h.authHandler.RenewToken(c, oldRefreshToken)
	c.JSON(status, resp)
}
