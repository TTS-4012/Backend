package api

import (
	"github.com/gin-gonic/gin"
	"ocontest/pkg"
	"strings"
)

const UserIDKey = "user_id"
const TokenTypeKey = "token_type"

func (h *handlers) Cors(c *gin.Context) {
	c.Header("access-control-allow-origin", "*")
	c.Next()
}

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
