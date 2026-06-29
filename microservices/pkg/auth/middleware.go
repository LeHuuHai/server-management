package authdomain

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

const BearerAuthScopes = "bearerAuth.Scopes"

// RoleCheckMiddleware checks if the user's role (from X-User-Role header)
// has the required scopes set by oapi-codegen.
func RoleCheckMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// 1. Get required scopes from gin context (injected by oapi-codegen wrapper)
		scopes, exists := c.Get(BearerAuthScopes)
		if !exists {
			// No security requirement for this endpoint
			c.Next()
			return
		}

		requiredScopes, ok := scopes.([]string)
		if !ok || len(requiredScopes) == 0 {
			// No specific scopes required
			c.Next()
			return
		}

		// 2. Extract user role from header (injected by Traefik API Gateway)
		roleStr := c.GetHeader("X-User-Role")
		if roleStr == "" {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"message": "missing user role header from gateway",
				"code":    "401",
			})
			return
		}

		userRole := Role(roleStr)
		userScopes := userRole.Scopes()

		// 3. Verify user has the required scopes
		for _, reqScope := range requiredScopes {
			hasScope := false
			for _, s := range userScopes {
				if string(s) == reqScope {
					hasScope = true
					break
				}
			}
			if !hasScope {
				c.AbortWithStatusJSON(http.StatusForbidden, gin.H{
					"message": "forbidden: missing required scope '" + reqScope + "'",
					"code":    "403",
				})
				return
			}
		}

		c.Next()
	}
}
