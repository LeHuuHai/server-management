package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/model"
	jwtprovider "github.com/LeHuuHai/server-management/microservices/pkg/jwt"
)

type AuthServiceInterface interface {
	Login(userName string, password string) (*model.LoginResult, error)
	HashPassword(password string) (string, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
	VerifyAccessToken(ctx context.Context, accessToken string) (*jwtprovider.AccessClaims, error)
}
