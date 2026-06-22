package handler

import (
	"context"
	"errors"

	"github.com/LeHuuHai/server-management/microservices/auth-service/api"
	serviceinterface "github.com/LeHuuHai/server-management/microservices/auth-service/internal/domain/service"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
)

type AuthHandler struct {
	authService serviceinterface.AuthServiceInterface
}

func NewAuthHandler(authService serviceinterface.AuthServiceInterface) *AuthHandler {
	return &AuthHandler{authService: authService}
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
	res, err := handler.authService.RefreshAccessToken(ctx, request.Body.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken),
			errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrRevokedToken):
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

// Logout
// (POST /auth/logout)
func (handler *AuthHandler) Logout(ctx context.Context, request api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	if err := handler.authService.Logout(ctx, request.Body.RefreshToken); err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken):
			return api.Logout401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		default:
			return api.Logout500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	return api.Logout200Response{}, nil
}
