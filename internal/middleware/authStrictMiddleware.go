package middleware

import (
	"log/slog"

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

func NewAuthStrictMiddleware(jwtProvider *jwtprovider.JWTProvider, blocklist *jwtprovider.TokenBlocklistRedis, reportKey string) api.StrictMiddlewareFunc {
	validToken := NewValidToken(jwtProvider, blocklist)
	validGetReportKey := NewAPIKeyMiddleware(reportKey)

	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(ctx *gin.Context, request interface{}) (interface{}, error) {
			if publicOps[operationID] {
				slog.Info("Public operation, no authentication required", "operationID", operationID)
				return f(ctx, request)
			}

			if getReportOps[operationID] {
				slog.Info("GetReport operation, validating report key", "operationID", operationID)
				validGetReportKey(ctx)
				if ctx.IsAborted() {
					return nil, nil
				}
				return f(ctx, request)
			}

			slog.Info("Protected operation, validating token", "operationID", operationID)
			// gọi ValidToken
			validToken(ctx)
			if ctx.IsAborted() {
				return nil, nil
			}
			slog.Info("Token valid, validating scope", "operationID", operationID)
			// gọi ValidScope
			ValidScope(ctx)
			if ctx.IsAborted() {
				return nil, nil
			}

			return f(ctx, request)
		}
	}
}
