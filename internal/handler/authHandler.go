package handler

import (
	"context"
	"errors"

	"github.com/LeHuuHai/server-management/api"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/service"
)

type AuthHandler struct {
	authService *service.AuthService
}

func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{
		authService: authService,
	}
}

// Login
// (POST /auth/login)
func (handler *AuthHandler) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	res, err := handler.authService.Login(request.Body.Username, request.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrInvalidCredentials):
			return api.Login401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		case errors.Is(err, apperr.ErrSignToken):
			return api.Login500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		default:
			return api.Login500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	return api.Login200JSONResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	}, nil
}

// Refresh token
// (POST /auth/refresh)
func (handler *AuthHandler) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	res, err := handler.authService.RefreshAccessToken(request.Body.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken),
			errors.Is(err, apperr.ErrRecordNotFound):
			return api.RefreshToken401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		case errors.Is(err, apperr.ErrSignToken):
			return api.RefreshToken500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		default:
			return api.RefreshToken500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	return api.RefreshToken200JSONResponse{
		AccessToken:  &res,
		RefreshToken: &request.Body.RefreshToken,
	}, nil
}
