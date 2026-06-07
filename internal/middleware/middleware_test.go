package middleware_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	masterconfig "github.com/LeHuuHai/server-management/config/master"
	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"github.com/LeHuuHai/server-management/internal/middleware"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/mocks"
	"github.com/gin-gonic/gin"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newGinContext(method, path string) (*gin.Context, *httptest.ResponseRecorder) {
	w := httptest.NewRecorder()
	c, _ := gin.CreateTestContext(w)
	c.Request = httptest.NewRequest(method, path, nil)
	return c, w
}

func newJWTProvider() *jwtprovider.JWTProvider {
	return jwtprovider.NewJWTProvider(&masterconfig.JWTConfig{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		AccessExpired:  3600,
		RefreshExpired: 86400,
	})
}

// ---------------------------------------------------------------------------
// NewAPIKeyMiddleware
// ---------------------------------------------------------------------------

func TestAPIKeyMiddleware_ValidKey_CallsNext(t *testing.T) {
	called := false
	handler := middleware.NewAPIKeyMiddleware("valid-key")

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("X-API-Key", "valid-key")

	router := gin.New()
	router.GET("/", handler, func(c *gin.Context) {
		called = true
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.True(t, called)
	assert.Equal(t, http.StatusOK, w.Code)
}

func TestAPIKeyMiddleware_InvalidKey_Returns401(t *testing.T) {
	handler := middleware.NewAPIKeyMiddleware("valid-key")

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("X-API-Key", "wrong-key")

	router := gin.New()
	router.GET("/", handler, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestAPIKeyMiddleware_MissingKey_Returns401(t *testing.T) {
	handler := middleware.NewAPIKeyMiddleware("valid-key")

	c, w := newGinContext(http.MethodGet, "/")
	// không set header

	router := gin.New()
	router.GET("/", handler, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// NewValidToken
// ---------------------------------------------------------------------------

func TestNewValidToken_MissingHeader_Returns401(t *testing.T) {
	jwt := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	c, w := newGinContext(http.MethodGet, "/")
	// không set Authorization header

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwt, mockBlocklist), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewValidToken_InvalidFormat_Returns401(t *testing.T) {
	jwt := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "InvalidFormat token")

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwt, mockBlocklist), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewValidToken_RevokedToken_Returns401(t *testing.T) {
	jwtP := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	// gen token hợp lệ
	token, err := jwtP.GenerateAccessToken(model.Account{UserID: 1, Role: authdomain.RoleUser})
	assert.NoError(t, err)

	mockBlocklist.EXPECT().IsRevoked(mock.Anything, token).Return(true, nil)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "Bearer "+token)

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwtP, mockBlocklist), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewValidToken_BlocklistError_Returns500(t *testing.T) {
	jwtP := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	token, err := jwtP.GenerateAccessToken(model.Account{UserID: 1, Role: authdomain.RoleUser})
	assert.NoError(t, err)

	mockBlocklist.EXPECT().IsRevoked(mock.Anything, token).Return(false, assert.AnError)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "Bearer "+token)

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwtP, mockBlocklist), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestNewValidToken_InvalidToken_Returns401(t *testing.T) {
	jwtP := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	mockBlocklist.EXPECT().IsRevoked(mock.Anything, "bad.token.string").Return(false, nil)

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "Bearer bad.token.string")

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwtP, mockBlocklist), func(c *gin.Context) {
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestNewValidToken_ValidToken_SetsContextAndCallsNext(t *testing.T) {
	jwtP := newJWTProvider()
	mockBlocklist := mocks.NewMockTokenBlocklist(t)

	token, err := jwtP.GenerateAccessToken(model.Account{UserID: 1, Role: authdomain.RoleAdmin})
	assert.NoError(t, err)

	mockBlocklist.EXPECT().IsRevoked(mock.Anything, token).Return(false, nil)

	var capturedUserID interface{}
	var capturedRole interface{}

	c, w := newGinContext(http.MethodGet, "/")
	c.Request.Header.Set("Authorization", "Bearer "+token)

	router := gin.New()
	router.GET("/", middleware.NewValidToken(jwtP, mockBlocklist), func(c *gin.Context) {
		capturedUserID, _ = c.Get("user_id")
		capturedRole, _ = c.Get("role")
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.Equal(t, uint(1), capturedUserID)
	assert.Equal(t, authdomain.RoleAdmin, capturedRole)
}

// ---------------------------------------------------------------------------
// ValidScope
// ---------------------------------------------------------------------------

func TestValidScope_NoScopeRequired_CallsNext(t *testing.T) {
	called := false
	c, w := newGinContext(http.MethodGet, "/")

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		// không set BearerAuthScopes → không yêu cầu scope
		middleware.ValidScope(c)
		called = true
		c.Status(http.StatusOK)
	})
	router.ServeHTTP(w, c.Request)

	assert.True(t, called)
}

func TestValidScope_HasRequiredScope_CallsNext(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Set(middleware.BearerAuthScopes, []string{"server:read"})
		c.Set("role", authdomain.RoleAdmin)
		middleware.ValidScope(c)
		if !c.IsAborted() {
			c.Status(http.StatusOK)
		}
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestValidScope_MissingScope_Returns403(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Set(middleware.BearerAuthScopes, []string{"user:read"})
		c.Set("role", authdomain.RoleGuest) // guest chỉ có server:read
		middleware.ValidScope(c)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusForbidden, w.Code)
}

func TestValidScope_MissingRoleInContext_Returns401(t *testing.T) {
	c, w := newGinContext(http.MethodGet, "/")

	router := gin.New()
	router.GET("/", func(c *gin.Context) {
		c.Set(middleware.BearerAuthScopes, []string{"server:read"})
		// không set role
		middleware.ValidScope(c)
	})
	router.ServeHTTP(w, c.Request)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

// ---------------------------------------------------------------------------
// LoggerFromContext
// ---------------------------------------------------------------------------

func TestLoggerFromContext_ReturnsDefault_WhenNotSet(t *testing.T) {
	logger := middleware.LoggerFromContext(context.Background())
	assert.NotNil(t, logger)
}
