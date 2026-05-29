package service

import (
	"fmt"

	"github.com/LeHuuHai/server-management/internal/domain/repo"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"golang.org/x/crypto/bcrypt"
)

type AuthService struct {
	jwtProvider *jwtprovider.JWTProvider
	accountRepo repo.AccountRepoInterface
}

func NewAuthService(
	jwtProvider *jwtprovider.JWTProvider,
	accountRepo repo.AccountRepoInterface,
) *AuthService {
	return &AuthService{
		jwtProvider: jwtProvider,
		accountRepo: accountRepo,
	}
}

type LoginResult struct {
	AccessToken  string `json:"access_token"`
	RefreshToken string `json:"refresh_token"`
}

func (s *AuthService) Login(userName string, password string) (*LoginResult, error) {
	account, err := s.accountRepo.FindByUserName(userName)
	if err != nil {
		return nil, fmt.Errorf("%w: %v", apperr.ErrRecordNotFound, err)
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

	return &LoginResult{
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

func (s *AuthService) RefreshAccessToken(refreshToken string) (string, error) {
	claims, err := s.jwtProvider.ParseRefreshToken(refreshToken)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrInvalidToken, err)
	}

	account, err := s.accountRepo.FindByUserID(claims.UserID)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrRecordNotFound, err)
	}

	token, err := s.jwtProvider.GenerateAccessToken(*account)
	if err != nil {
		return "", fmt.Errorf("%w: %v", apperr.ErrSignToken, err)
	}
	return token, nil
}
