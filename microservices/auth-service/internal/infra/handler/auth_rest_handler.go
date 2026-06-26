package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LeHuuHai/server-management/microservices/auth-service/api"
	serviceinterface "github.com/LeHuuHai/server-management/microservices/auth-service/internal/domain/service"
	apperr "github.com/LeHuuHai/server-management/microservices/pkg/apperr"
)

type AuthRestHandler struct {
	authService serviceinterface.AuthServiceInterface
}

func NewAuthRestHandler(authService serviceinterface.AuthServiceInterface) *AuthRestHandler {
	return &AuthRestHandler{
		authService: authService,
	}
}

func strPtr(s string) *string {
	return &s
}

// Login
// (POST /auth/login)
func (handler *AuthRestHandler) Login(ctx context.Context, request api.LoginRequestObject) (api.LoginResponseObject, error) {
	slog.Info("handler: login", slog.String("username", request.Body.Username))

	res, err := handler.authService.Login(request.Body.Username, request.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrInvalidCredentials):
			slog.Warn("invalid credentials", slog.String("username", request.Body.Username))
			return api.Login401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		default:
			slog.Error("failed to login", slog.Any("err", err))
			return api.Login500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	slog.Info("handler: login success", slog.String("username", request.Body.Username))

	return api.Login200JSONResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	}, nil
}

// Refresh token
// (POST /auth/refresh)
func (handler *AuthRestHandler) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	slog.Info("handler: refresh token")

	res, err := handler.authService.RefreshAccessToken(ctx, request.Body.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken),
			errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrRevokedToken):
			slog.Warn("invalid refresh token", slog.Any("err", err))
			return api.RefreshToken401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		default:
			slog.Error("failed to refresh token", slog.Any("err", err))
			return api.RefreshToken500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	slog.Info("handler: refresh token success")

	return api.RefreshToken200JSONResponse{
		AccessToken:  &res,
		RefreshToken: &request.Body.RefreshToken,
	}, nil
}

// Logout
// (POST /auth/logout)
func (handler *AuthRestHandler) Logout(ctx context.Context, request api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	slog.Info("handler: logout")

	if err := handler.authService.Logout(ctx, request.Body.RefreshToken); err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken):
			slog.Warn("invalid token", slog.Any("err", err))
			return api.Logout401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		default:
			slog.Error("failed to logout", slog.Any("err", err))
			return api.Logout500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	slog.Info("handler: logout success")

	return api.Logout200Response{}, nil
}
