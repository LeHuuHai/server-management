package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/LeHuuHai/server-management/api"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

type contextKey string

const loggerKey contextKey = "logger"

func NewLogStrictMiddleware(logger *slog.Logger) api.StrictMiddlewareFunc {
	return func(f api.StrictHandlerFunc, operationID string) api.StrictHandlerFunc {
		return func(c *gin.Context, request interface{}) (interface{}, error) {
			start := time.Now()

			// Gắn logger vào context
			requestID := uuid.New().String()
			reqLogger := logger.With(
				slog.String("request_id", requestID),
				slog.String("operation_id", operationID),
				slog.String("method", c.Request.Method),
				slog.String("path", c.FullPath()),
			)
			// gin.Context
			c.Set("logger", reqLogger)
			// context.Context
			ctx := context.WithValue(c.Request.Context(), loggerKey, reqLogger)
			c.Request = c.Request.WithContext(ctx)

			reqLogger.Info("request started")

			res, err := f(c, request)

			reqLogger.Info("request completed",
				slog.Int("status", c.Writer.Status()),
				slog.Duration("duration", time.Since(start)),
			)

			return res, err
		}
	}
}
