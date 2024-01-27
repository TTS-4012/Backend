package api

import (
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
	"github.com/ocontest/backend/pkg"
	"github.com/ocontest/backend/pkg/structs"
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
		resp, status = h.authHandler.LoginWithPassword(c, reqData.Email, reqData.Password)
	case "otp":
		resp, status = h.authHandler.LoginWithOTP(c, reqData.Email, reqData.OTP)
	default:
		logger.Warning("invalid grant type! reqData: ", reqData)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "grant type must be either 'password' or 'otp'",
		})
		return
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

	status := h.authHandler.RequestLoginWithOTP(c, reqData.Email)
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

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}
	reqData.UserID = userID.(int64)

	status := h.authHandler.EditUser(c, reqData)
	c.Status(status)
}

func (h *handlers) getOwnUser(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "getOwnUser")

	userID, exists := c.Get(UserIDKey)
	if !exists {
		logger.Error("error on getting user_id from context")
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": pkg.ErrInternalServerError.Error(),
		})
		return
	}

	resp, status := h.authHandler.GetUser(c, userID.(int64), true)
	c.JSON(status, resp)
}

func (h *handlers) getUser(c *gin.Context) {
	logger := pkg.Log.WithField("handler", "getUser")

	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		logger.Error("error on getting user_id from url: ", err)
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "invalid user_id, user_id should be an integer",
		})
		return
	}

	resp, status := h.authHandler.GetUser(c, userID, false)
	c.JSON(status, resp)
}
