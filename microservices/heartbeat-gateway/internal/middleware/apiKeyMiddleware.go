package middleware

import (
	"log/slog"
	"net/http"

	"github.com/gin-gonic/gin"
)

func NewAPIKeyMiddleware(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != validKey {
			slog.Info("Invalid API key provided", "providedKey", key)
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"code":    "UNAUTHORIZED",
				"message": "invalid api key",
			})
			return
		}
		slog.Debug("API key is valid", "providedKey", key)
		c.Next()
	}
}
