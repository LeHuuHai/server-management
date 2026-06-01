package middleware

import (
	"github.com/LeHuuHai/server-management/api"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"github.com/gin-gonic/gin"
)

var publicOps = map[string]bool{
	"Login":        true,
	"RefreshToken": true,
}

var getReportOps = map[string]bool{
	"GetReportFile": true,
}

func NewAuthStrictMiddleware(jwtProvider *jwtprovider.JWTProvider, reportKey string) api.StrictMiddlewareFunc {
	validToken := NewValidToken(jwtProvider)
	validGetReportKey := MewAPIKeyMiddleware(reportKey)

	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx *gin.Context, request interface{}) (interface{}, error) {
			if publicOps[operationID] {
				return f(ctx, request)
			}

			if getReportOps[operationID] {
				validGetReportKey(ctx)
				if ctx.IsAborted() {
					return nil, nil
				}
				return f(ctx, request)
			}

			// gọi ValidToken
			validToken(ctx)
			if ctx.IsAborted() {
				return nil, nil
			}

			// gọi ValidScope
			ValidScope(ctx)
			if ctx.IsAborted() {
				return nil, nil
			}

			return f(ctx, request)
		}
	}
}
