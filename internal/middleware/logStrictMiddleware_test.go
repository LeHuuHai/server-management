package middleware_test

import (
	"bytes"
	"context"
	"log/slog"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/LeHuuHai/server-management/api"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestLogger(buf *bytes.Buffer) *slog.Logger {
	return slog.New(slog.NewTextHandler(buf, &slog.HandlerOptions{Level: slog.LevelDebug}))
}

// ---------------------------------------------------------------------------
// NewLogStrictMiddleware
// ---------------------------------------------------------------------------

func TestLogStrictMiddleware_LogsRequestStartedAndCompleted(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)

	// handler giả — trả về 200
	handlerCalled := false
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		handlerCalled = true
		c.Status(http.StatusOK)
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "TestOperation")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/test", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	assert.True(t, handlerCalled)

	logs := buf.String()
	assert.Contains(t, logs, "request started")
	assert.Contains(t, logs, "request completed")
	assert.Contains(t, logs, "TestOperation")
}

func TestLogStrictMiddleware_LogsContainRequestID(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodPost, "/servers", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	assert.Contains(t, buf.String(), "request_id")
}

func TestLogStrictMiddleware_SetsLoggerInGinContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)

	var capturedLogger interface{}
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		capturedLogger, _ = c.Get("logger")
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	assert.NotNil(t, capturedLogger)
	_, ok := capturedLogger.(*slog.Logger)
	assert.True(t, ok)
}

func TestLogStrictMiddleware_SetsLoggerInRequestContext(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)

	var capturedCtx context.Context
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		capturedCtx = c.Request.Context()
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	assert.NotNil(t, capturedCtx)
	// logger được inject vào context
	loggerFromCtx := middleware.LoggerFromContext(capturedCtx)
	assert.NotNil(t, loggerFromCtx)
}

func TestLogStrictMiddleware_LogsMethodAndPath(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodDelete, "/servers/s1", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	logs := buf.String()
	assert.Contains(t, logs, "DELETE")
}

func TestLogStrictMiddleware_PropagatesHandlerError(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		return nil, assert.AnError
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := wrappedHandler(c, nil)

	assert.ErrorIs(t, err, assert.AnError)
	// request completed vẫn được log dù có lỗi
	assert.Contains(t, buf.String(), "request completed")
}

func TestLogStrictMiddleware_LogsDuration(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")

	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)

	_, err := wrappedHandler(c, nil)

	require.NoError(t, err)
	assert.True(t, strings.Contains(buf.String(), "duration"))
}

// ---------------------------------------------------------------------------
// LoggerFromContext
// ---------------------------------------------------------------------------

func TestLoggerFromContext_ReturnsInjectedLogger(t *testing.T) {
	buf := &bytes.Buffer{}
	logger := newTestLogger(buf)

	mw := middleware.NewLogStrictMiddleware(logger)

	var capturedCtx context.Context
	innerHandler := func(c *gin.Context, request interface{}) (interface{}, error) {
		capturedCtx = c.Request.Context()
		return nil, nil
	}

	wrappedHandler := mw(innerHandler, "op-id")
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(http.MethodGet, "/", nil)
	wrappedHandler(c, nil)

	result := middleware.LoggerFromContext(capturedCtx)
	assert.NotNil(t, result)
}

func TestLoggerFromContext_ReturnsDefaultWhenNotSet(t *testing.T) {
	result := middleware.LoggerFromContext(context.Background())
	assert.NotNil(t, result)
}

// ---------------------------------------------------------------------------
// Ensure NewLogStrictMiddleware returns api.StrictMiddlewareFunc
// ---------------------------------------------------------------------------

func TestLogStrictMiddleware_ImplementsStrictMiddlewareFunc(t *testing.T) {
	logger := slog.Default()
	var mw api.StrictMiddlewareFunc = middleware.NewLogStrictMiddleware(logger)
	assert.NotNil(t, mw)
}
