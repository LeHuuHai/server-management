package serviceinterface

import (
	"context"

	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/model"
)

type AuthServiceInterface interface {
	Login(userName string, password string) (*model.LoginResult, error)
	HashPassword(password string) (string, error)
	RefreshAccessToken(ctx context.Context, refreshToken string) (string, error)
	Logout(ctx context.Context, refreshToken string) error
}
