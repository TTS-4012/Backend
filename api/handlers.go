package api

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"ocontest/internal/oc/auth"
	"ocontest/pkg"
	"ocontest/pkg/structs"
)

type handlers struct {
	processor auth.AuthHandler
}

func AddRoutes(r *gin.Engine, authHandler auth.AuthHandler) {
	h := handlers{authHandler}

	v1 := r.Group("/v1")
	{
		authGroup := v1.Group("/auth")
		{
			authGroup.POST("/register_user", h.registerUser)
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

	resp, status, err := h.processor.RegisterUser(c, reqData)
	if err != nil {
		pkg.Log.Error("Failed to register user", err)
	}
	c.JSON(status, resp)
}

func (h *handlers) loginUser(c *gin.Context) {
	panic("implement me")
}

func (h *handlers) renewToken(c *gin.Context) {
	panic("implement me")
}
