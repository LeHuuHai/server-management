package middleware

import (
	"log/slog"
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
			slog.Info("Missing or invalid Authorization header", "header", authHeader)
			c.AbortWithStatusJSON(401, gin.H{"error": "missing or invalid token format"})
			return
		}

		tokenString := strings.TrimPrefix(authHeader, "Bearer ")
		slog.Info("Validating token", "token", tokenString)

		claims, err := jwtProvider.ParseAccessToken(tokenString)
		if err != nil {
			slog.Error("Error occurred while parsing token", "error", err)
			c.AbortWithStatusJSON(401, gin.H{"error": "invalid token"})
			return
		}
		slog.Info("Token parsed", "user_id", claims.UserID, "role", claims.Role, "expires_at", claims.ExpiresAt)

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
	requiredScopes := requiredVal.([]string)

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

	roleScopes := make(map[string]bool)
	for _, scope := range role.Scopes() {
		roleScopes[string(scope)] = true
	}

	for _, required := range requiredScopes {
		if !roleScopes[required] {
			slog.Info("User role does not have required scope", "requiredScope", required)
			c.AbortWithStatusJSON(403, gin.H{"error": "forbidden"})
			return
		}
	}

	slog.Info("All required scopes are present", "requiredScopes", requiredScopes)
	c.Next()
}
