package middleware

import "github.com/gin-gonic/gin"

func MewAPIKeyMiddleware(validKey string) gin.HandlerFunc {
	return func(c *gin.Context) {
		key := c.GetHeader("X-API-Key")
		if key != validKey {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid api key"})
			return
		}
		c.Next()
	}
}
