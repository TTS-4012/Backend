package api

import (
	"github.com/gin-gonic/gin"
	"ocontest/pkg"
	"strings"
)

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

		c.Set("user_id", userId)
		c.Set("token_type", typ)

		c.Next()
	}
}
