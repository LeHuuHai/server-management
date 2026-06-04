package middleware

import (
	"context"
	"log/slog"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func LogMiddleware(logger *slog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := uuid.New().String()
		reqLogger := logger.With(
			slog.String("request_id", requestID),
			slog.String("method", c.Request.Method),
			slog.String("path", c.FullPath()),
		)

		// Gắn logger vào cả gin.Context lẫn context.Context
		c.Set("logger", reqLogger)

		ctx := context.WithValue(c.Request.Context(), "logger", reqLogger)
		c.Request = c.Request.WithContext(ctx)

		reqLogger.Info("request started")
		c.Next()
		reqLogger.Info("request completed",
			slog.Int("status", c.Writer.Status()),
			slog.Duration("duration", time.Since(start)),
		)
	}
}
