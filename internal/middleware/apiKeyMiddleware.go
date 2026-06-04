package middleware

import (
	"log/slog"

	"github.com/gin-gonic/gin"
)

func NewAPIKeyMiddleware(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != validKey {
			slog.Info("Invalid API key provided", "providedKey", key)
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
			return
		}
		slog.Info("API key is valid", "providedKey", key)
		c.Next()
	}
}
