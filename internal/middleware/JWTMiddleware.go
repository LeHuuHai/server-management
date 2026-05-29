package middleware

import (
	"strings"

	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"github.com/gin-gonic/gin"
)

const (
	BearerAuthScopes = "bearerAuth.Scopes"
)

func NewValidToken(jwtProvider *jwtprovider.JWTProvider) gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" || !strings.HasPrefix(authHeader, "Bearer ") {
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid token format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")

		claims, err := jwtProvider.ParseAccessToken(tokenString)
		if err != nil {
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}

		c.Set("user_id", claims.UserID)
		c.Set("role", claims.Role)

		c.Next()
	}
}

func ValidScope(c *gin.Context) {
	requiredVal, exists := c.Get(BearerAuthScopes)
	if !exists {
		c.Next()
		return
	}
	required := requiredVal.(authdomain.Scope)

	roleVal, ok := c.Get("role")
	if !ok {
		c.AbortWithStatusJSON(401, gin.H{"error": "missing role"})
		return
	}

	role, ok := roleVal.(authdomain.Role)
	if !ok {
		c.AbortWithStatusJSON(500, gin.H{"error": "invalid role"})
		return
	}

	for _, scope := range role.Scopes() {
		if scope == required {
			c.Next()
			return
		}
	}

	c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
}
