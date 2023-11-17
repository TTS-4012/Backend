package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

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

	var reqData structs.RequestVerifyEmail
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
	var reqData structs.RequestLogin

	if err := c.ShouldBindJSON(&reqData); err != nil {
		logger.Error("error on binding request data json")
		c.JSON(http.StatusBadRequest, gin.H{
			"error": pkg.ErrBadRequest.Error(),
		})
		return
	}

	var resp structs.AuthenticateResponse
	var status int
	switch reqData.GrantType {
	case "password":
		resp, status = h.authHandler.LoginWithPassword(c, reqData.UserName, reqData.Password)
	case "otp":
		resp, status = h.authHandler.LoginWithOTP(c, reqData.UserID, reqData.OTP)
	}
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

func (h *handlers) GetOTPForLogin(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "GetOTPForLogin")

	var reqData structs.RequestGetOTPLogin
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
