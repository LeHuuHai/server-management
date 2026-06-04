package jwtprovider_test

import (
	"testing"
	"time"

	masterconfig "github.com/LeHuuHai/server-management/config/master"
	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

func newProvider(accessExpire, refreshExpire int) *jwtprovider.JWTProvider {
	cfg := &masterconfig.JWTConfig{
		AccessSecret:   "access-secret-key",
		RefreshSecret:  "refresh-secret-key",
		AccessExpired:  accessExpire,
		RefreshExpired: refreshExpire,
	}
	return jwtprovider.NewJWTProvider(cfg)
}

var testAccount = model.Account{
	ID:       1,
	UserID:   42,
	Username: "alice",
	Role:     authdomain.RoleAdmin,
}

func TestGenerateAndParseAccessToken(t *testing.T) {
	p := newProvider(3600, 86400)

	token, err := p.GenerateAccessToken(testAccount)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	if token == "" {
		t.Fatal("expected non-empty token")
	}

	claims, err := p.ParseAccessToken(token)
	if err != nil {
		t.Fatalf("ParseAccessToken error: %v", err)
	}
	if claims.UserID != testAccount.UserID {
		t.Errorf("expected UserID %d, got %d", testAccount.UserID, claims.UserID)
	}
	if claims.Role != authdomain.RoleAdmin {
		t.Errorf("expected role admin, got %s", claims.Role)
	}
	if claims.TokenType != jwtprovider.TokenTypeAccess {
		t.Errorf("expected token type access, got %s", claims.TokenType)
	}
	if claims.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected issued_at to be set")
	}
}

func TestGenerateAndParseRefreshToken(t *testing.T) {
	p := newProvider(3600, 86400)

	token, err := p.GenerateRefreshToken(testAccount)
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	claims, err := p.ParseRefreshToken(token)
	if err != nil {
		t.Fatalf("ParseRefreshToken error: %v", err)
	}
	if claims.UserID != testAccount.UserID {
		t.Errorf("expected UserID %d, got %d", testAccount.UserID, claims.UserID)
	}
	if claims.TokenType != jwtprovider.TokenTypeRefresh {
		t.Errorf("expected token type refresh, got %s", claims.TokenType)
	}
	if claims.ExpiresAt == nil {
		t.Error("expected expires_at to be set")
	}
	if claims.IssuedAt == nil {
		t.Error("expected issued_at to be set")
	}
}

func TestParseAccessToken_WrongSecret(t *testing.T) {
	p1 := newProvider(3600, 86400)
	// Create a provider with a different secret
	cfg2 := &masterconfig.JWTConfig{
		AccessSecret:   "different-secret",
		RefreshSecret:  "different-refresh",
		AccessExpired:  3600,
		RefreshExpired: 86400,
	}
	p2 := jwtprovider.NewJWTProvider(cfg2)

	token, err := p1.GenerateAccessToken(testAccount)
	if err != nil {
		t.Fatalf("GenerateAccessToken error: %v", err)
	}
	_, err = p2.ParseAccessToken(token)
	if err == nil {
		t.Error("expected error when parsing token with wrong secret")
	}
}

func TestParseRefreshToken_WrongSecret(t *testing.T) {
	p1 := newProvider(3600, 86400)
	cfg2 := &masterconfig.JWTConfig{
		AccessSecret:   "wrong",
		RefreshSecret:  "wrong-refresh",
		AccessExpired:  3600,
		RefreshExpired: 86400,
	}
	p2 := jwtprovider.NewJWTProvider(cfg2)

	token, err := p1.GenerateRefreshToken(testAccount)
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}
	_, err = p2.ParseRefreshToken(token)
	if err == nil {
		t.Error("expected error when parsing refresh token with wrong secret")
	}
}

func TestParseAccessToken_Expired(t *testing.T) {
	p := newProvider(-1, 86400) // -1 second = already expired

	token, err := p.GenerateAccessToken(testAccount)
	if err != nil {
		t.Fatalf("generate error: %v", err)
	}

	// Give it a moment to actually expire
	time.Sleep(10 * time.Millisecond)

	_, err = p.ParseAccessToken(token)
	if err == nil {
		t.Error("expected error for expired access token")
	}
}

func TestParseRefreshToken_Expired(t *testing.T) {
	p := newProvider(3600, -1)

	token, err := p.GenerateRefreshToken(testAccount)
	if err != nil {
		t.Fatalf("GenerateRefreshToken error: %v", err)
	}

	_, err = p.ParseRefreshToken(token)
	if err == nil {
		t.Fatal("expected error for expired refresh token")
	}
}

func TestParseAccessToken_InvalidString(t *testing.T) {
	p := newProvider(3600, 86400)
	_, err := p.ParseAccessToken("this.is.not.valid")
	if err == nil {
		t.Error("expected error for garbage token string")
	}
}

func TestParseRefreshToken_InvalidString(t *testing.T) {
	p := newProvider(3600, 86400)
	_, err := p.ParseRefreshToken("this.is.not.valid")
	if err == nil {
		t.Error("expected error for garbage token string")
	}
}

func TestAccessToken_ContainsRole(t *testing.T) {
	p := newProvider(3600, 86400)
	for _, role := range []authdomain.Role{authdomain.RoleAdmin, authdomain.RoleUser, authdomain.RoleGuest} {
		acc := model.Account{UserID: 1, Role: role}
		token, err := p.GenerateAccessToken(acc)
		if err != nil {
			t.Fatalf("generate error for role %s: %v", role, err)
		}
		claims, err := p.ParseAccessToken(token)
		if err != nil {
			t.Fatalf("parse error for role %s: %v", role, err)
		}
		if claims.Role != role {
			t.Errorf("expected role %s, got %s", role, claims.Role)
		}
	}
}

func TestParseAccessToken_WrongSigningMethod(t *testing.T) {
	p := newProvider(3600, 86400)

	claims := jwtprovider.AccessClaims{
		UserID:    testAccount.UserID,
		Role:      testAccount.Role,
		TokenType: jwtprovider.TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).
		SignedString([]byte("access-secret-key"))
	if err != nil {
		t.Fatalf("SignedString error: %v", err)
	}

	_, err = p.ParseAccessToken(token)
	if err == nil {
		t.Fatal("expected error for wrong signing method")
	}
}

func TestParseRefreshToken_WrongSigningMethod(t *testing.T) {
	p := newProvider(3600, 86400)

	claims := jwtprovider.RefreshClaims{
		UserID:    testAccount.UserID,
		TokenType: jwtprovider.TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(24 * time.Hour)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token, err := jwt.NewWithClaims(jwt.SigningMethodHS512, claims).
		SignedString([]byte("refresh-secret-key"))
	if err != nil {
		t.Fatalf("SignedString error: %v", err)
	}

	_, err = p.ParseRefreshToken(token)
	if err == nil {
		t.Fatal("expected error for wrong signing method")
	}
}
