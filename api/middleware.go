package api

import (
	"github.com/gin-gonic/gin"
	"ocontest/pkg"
	"strings"
)

const UserIDKey = "user_id"
const TokenTypeKey = "token_type"

func (h *handlers) AuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		logger := pkg.Log.WithField("middleware", "Auth")

		authHeader := c.GetHeader("Authorization")
		authHeader = strings.Replace(authHeader, "Bearer ", "", 1)
		userId, typ, err := h.authHandler.ParseAuthToken(c, authHeader)

		if err != nil {
			logger.Error("error on parsing token", err)
			c.AbortWithStatusJSON(401, gin.H{"message": "invalid token"})
			return
		}

		c.Set(UserIDKey, userId)
		c.Set(TokenTypeKey, typ)

		c.Next()
	}
}

func (h *handlers) corsHandler(c *gin.Context) {
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE, UPDATE")
	c.Header("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, full-refresh")

	c.Next()
}
