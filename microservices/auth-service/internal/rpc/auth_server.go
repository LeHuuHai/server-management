package rpc

import (
	"context"

	serviceinterface "github.com/LeHuuHai/server-management/microservices/auth-service/internal/domain/service"
	authpb "github.com/LeHuuHai/server-management/microservices/pkg/pb/auth"
)

type AuthServer struct {
	authpb.UnimplementedAuthServiceServer
	authService serviceinterface.AuthServiceInterface
}

func NewAuthServer(authService serviceinterface.AuthServiceInterface) *AuthServer {
	return &AuthServer{
		authService: authService,
	}
}

func (s *AuthServer) Login(ctx context.Context, req *authpb.LoginRequest) (*authpb.LoginResponse, error) {
	result, err := s.authService.Login(req.GetUsername(), req.GetPassword())
	if err != nil {
		return nil, err
	}

	return &authpb.LoginResponse{
		AccessToken:  result.AccessToken,
		RefreshToken: result.RefreshToken,
	}, nil
}

func (s *AuthServer) RefreshToken(ctx context.Context, req *authpb.RefreshTokenRequest) (*authpb.RefreshTokenResponse, error) {
	token, err := s.authService.RefreshAccessToken(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &authpb.RefreshTokenResponse{
		AccessToken:  token,
		RefreshToken: req.GetRefreshToken(), // Keep same refresh token
	}, nil
}

func (s *AuthServer) Logout(ctx context.Context, req *authpb.LogoutRequest) (*authpb.LogoutResponse, error) {
	err := s.authService.Logout(ctx, req.GetRefreshToken())
	if err != nil {
		return nil, err
	}

	return &authpb.LogoutResponse{
		Success: true,
	}, nil
}
