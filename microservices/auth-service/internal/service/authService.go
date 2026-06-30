package service

import (
	"context"
	"fmt"

	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/domain/cache"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/domain/repo"
	"github.com/LeHuuHai/server-management/microservices/auth-service/internal/model"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
	jwtprovider "github.com/LeHuuHai/server-management/microservices/pkg/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtProvider *jwtprovider.JWTProvider
	blocklist   cache.TokenBlocklist
	accountRepo repo.AccountRepoInterface
}

func NewAuthService(
	jwtProvider *jwtprovider.JWTProvider,
	blocklist cache.TokenBlocklist,
	accountRepo repo.AccountRepoInterface,
) *AuthService {
	return &AuthService{
		jwtProvider: jwtProvider,
		blocklist:   blocklist,
		accountRepo: accountRepo,
	}
}

func (s *AuthService) Login(userName string, password string) (*model.LoginResult, error) {
	account, err := s.accountRepo.FindByUserName(userName)
	if err != nil {
		return nil, err
	}

	err = bcrypt.CompareHashAndPassword(
		[]byte(account.Password),
		[]byte(password),
	)

	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrInvalidCredentials, err)
	}

	accessToken, err := s.jwtProvider.GenerateAccessToken(*account)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrSignToken, err)
	}

	refreshToken, err := s.jwtProvider.GenerateRefreshToken(*account)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrSignToken, err)
	}

	return &model.LoginResult{
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

func (s *AuthService) HashPassword(password string) (string, error) {
	hash, err := bcrypt.GenerateFromPassword(
		[]byte(password),
		bcrypt.DefaultCost,
	)
	if err != nil {
		return "", err
	}

	return string(hash), nil
}

func (s *AuthService) RefreshAccessToken(ctx context.Context, refreshToken string) (string, error) {
	revoked, err := s.blocklist.IsRevoked(ctx, refreshToken)
	if err != nil {
		return "", fmt.Errorf("refresh: %w", err)
	}
	if revoked {
		return "", fmt.Errorf("%w: token has been revoked", apperr.ErrRevokedToken)
	}
	claims, err := s.jwtProvider.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrInvalidToken, err)
	}

	account, err := s.accountRepo.FindByUserID(claims.UserID)
	if err != nil {
		return "", err
	}

	token, err := s.jwtProvider.GenerateAccessToken(*account)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrSignToken, err)
	}
	return token, nil
}

func (s *AuthService) Logout(ctx context.Context, refreshToken string) error {
	claims, err := s.jwtProvider.ParseRefreshToken(refreshToken)
	if err != nil {
		return fmt.Errorf("%w: %v", apperr.ErrInvalidToken, err)
	}
	if err := s.blocklist.Revoke(ctx, refreshToken, claims.ExpiresAt.Time); err != nil {
		return fmt.Errorf("logout: %w", err)
	}
	return nil
}

func (s *AuthService) VerifyAccessToken(ctx context.Context, accessToken string) (*jwtprovider.AccessClaims, error) {
	claims, err := s.jwtProvider.ParseAccessToken(accessToken)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrInvalidToken, err)
	}

	return claims, nil
}

