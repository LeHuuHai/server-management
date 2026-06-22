package jwtprovider

import (
	"errors"
	"time"

	authdomain "github.com/LeHuuHai/server-management/microservices/pkg/auth"
	commonconfig "github.com/LeHuuHai/server-management/microservices/pkg/config"
	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	accessSecret  []byte
	refreshSecret []byte
	accessExpire  int64
	refreshExpire int64
}

const (
	TokenTypeAccess  = "access"
	TokenTypeRefresh = "refresh"
)

type AccountClaimsSource interface {
	GetUserID() uint
	GetRole() authdomain.Role
}

func NewJWTProvider(cfg *commonconfig.JWTConfig) *JWTProvider {
	return &JWTProvider{
		accessSecret:  []byte(cfg.AccessSecret),
		refreshSecret: []byte(cfg.RefreshSecret),
		accessExpire:  int64(cfg.AccessExpired),
		refreshExpire: int64(cfg.RefreshExpired),
	}
}

type AccessClaims struct {
	UserID    uint            `json:"user_id"`
	Role      authdomain.Role `json:"role"`
	TokenType string          `json:"token_type"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID    uint   `json:"user_id"`
	TokenType string `json:"token_type"`
	jwt.RegisteredClaims
}

func validateSigningMethod(token *jwt.Token) error {
	if token.Method != jwt.SigningMethodHS256 {
		return errors.New("unexpected signing method")
	}
	return nil
}

func (s *JWTProvider) GenerateAccessToken(account AccountClaimsSource) (string, error) {
	claims := AccessClaims{
		UserID:    account.GetUserID(),
		Role:      account.GetRole(),
		TokenType: TokenTypeAccess,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.accessExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.accessSecret)
}

func (s *JWTProvider) GenerateRefreshToken(account AccountClaimsSource) (string, error) {
	claims := RefreshClaims{
		UserID:    account.GetUserID(),
		TokenType: TokenTypeRefresh,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Duration(s.refreshExpire) * time.Second)),
			IssuedAt:  jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	return token.SignedString(s.refreshSecret)
}

func (s *JWTProvider) ParseAccessToken(tokenString string) (*AccessClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &AccessClaims{}, func(token *jwt.Token) (interface{}, error) {
		if err := validateSigningMethod(token); err != nil {
			return nil, err
		}
		return s.accessSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}
	if claims.TokenType != TokenTypeAccess {
		return nil, errors.New("invalid access token type")
	}
	return claims, nil
}

func (s *JWTProvider) ParseRefreshToken(tokenString string) (*RefreshClaims, error) {
	token, err := jwt.ParseWithClaims(tokenString, &RefreshClaims{}, func(token *jwt.Token) (interface{}, error) {
		if err := validateSigningMethod(token); err != nil {
			return nil, err
		}
		return s.refreshSecret, nil
	})
	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}
	if claims.TokenType != TokenTypeRefresh {
		return nil, errors.New("invalid refresh token type")
	}
	return claims, nil
}
