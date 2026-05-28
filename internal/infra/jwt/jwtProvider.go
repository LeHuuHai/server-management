package jwtprovider

import (
	"errors"
	"time"

	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/golang-jwt/jwt/v5"
)

type JWTProvider struct {
	accessSecret  []byte
	refreshSecret []byte
	accesExpire   int64
	refreshExpire int64
}

func NewJWTProvider(accessToken string, refreshToken string, accessExpire int64, refreshExpire int64) *JWTProvider {
	return &JWTProvider{
		accessSecret:  []byte(accessToken),
		refreshSecret: []byte(refreshToken),
		accesExpire:   accessExpire,
		refreshExpire: refreshExpire,
	}
}

type AccessClaims struct {
	UserID uint            `json:"user_id"`
	Role   authdomain.Role `json:"role"`
	jwt.RegisteredClaims
}

type RefreshClaims struct {
	UserID uint `json:"user_id"`
	jwt.RegisteredClaims
}

func (s *JWTProvider) GenerateAccessToken(
	account model.Account,
) (string, error) {

	claims := AccessClaims{
		UserID: account.UserID,
		Role:   account.Role,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Duration(s.accesExpire) * time.Second),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.accessSecret)
}

func (s *JWTProvider) GenerateRefreshToken(
	account model.Account,
) (string, error) {

	claims := RefreshClaims{
		UserID: account.UserID,

		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(
				time.Now().Add(time.Duration(s.refreshExpire) * time.Second),
			),
			IssuedAt: jwt.NewNumericDate(time.Now()),
		},
	}

	token := jwt.NewWithClaims(
		jwt.SigningMethodHS256,
		claims,
	)

	return token.SignedString(s.refreshSecret)
}

func (s *JWTProvider) ParseAccessToken(
	tokenString string,
) (*AccessClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&AccessClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.accessSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*AccessClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid access token")
	}

	return claims, nil
}

func (s *JWTProvider) ParseRefreshToken(
	tokenString string,
) (*RefreshClaims, error) {

	token, err := jwt.ParseWithClaims(
		tokenString,
		&RefreshClaims{},
		func(token *jwt.Token) (interface{}, error) {
			return s.refreshSecret, nil
		},
	)

	if err != nil {
		return nil, err
	}

	claims, ok := token.Claims.(*RefreshClaims)
	if !ok || !token.Valid {
		return nil, errors.New("invalid refresh token")
	}

	return claims, nil
}
