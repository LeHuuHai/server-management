package handler

import (
	"context"
	"errors"
	"log/slog"

	"github.com/LeHuuHai/server-management/api"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/middleware"
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
	logger := middleware.LoggerFromContext(ctx)

	logger.Info("handler: login", slog.String("username", request.Body.Username))

	res, err := handler.authService.Login(request.Body.Username, request.Body.Password)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrInvalidCredentials):
			logger.Warn("invalid credentials", slog.String("username", request.Body.Username))
			return api.Login401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		case errors.Is(err, apperr.ErrSignToken):
			logger.Error("failed to sign token", slog.Any("err", err))
			return api.Login500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		default:
			logger.Error("failed to login", slog.Any("err", err))
			return api.Login500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	logger.Info("handler: login success", slog.String("username", request.Body.Username))

	return api.Login200JSONResponse{
		AccessToken:  &res.AccessToken,
		RefreshToken: &res.RefreshToken,
	}, nil
}

// Refresh token
// (POST /auth/refresh)
func (handler *AuthHandler) RefreshToken(ctx context.Context, request api.RefreshTokenRequestObject) (api.RefreshTokenResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	logger.Info("handler: refresh token")

	res, err := handler.authService.RefreshAccessToken(ctx, request.Body.RefreshToken)
	if err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken),
			errors.Is(err, apperr.ErrRecordNotFound),
			errors.Is(err, apperr.ErrRevokedToken):
			logger.Warn("invalid refresh token", slog.Any("err", err))
			return api.RefreshToken401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		case errors.Is(err, apperr.ErrSignToken):
			logger.Error("failed to sign token", slog.Any("err", err))
			return api.RefreshToken500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		default:
			logger.Error("failed to refresh token", slog.Any("err", err))
			return api.RefreshToken500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	logger.Info("handler: refresh token success")

	return api.RefreshToken200JSONResponse{
		AccessToken:  &res,
		RefreshToken: &request.Body.RefreshToken,
	}, nil
}

// Logout
// (POST /auth/logout)
func (handler *AuthHandler) Logout(ctx context.Context, request api.LogoutRequestObject) (api.LogoutResponseObject, error) {
	logger := middleware.LoggerFromContext(ctx)

	logger.Info("handler: logout")

	if err := handler.authService.Logout(ctx, request.Body.RefreshToken); err != nil {
		switch {
		case errors.Is(err, apperr.ErrInvalidToken):
			logger.Warn("invalid token", slog.Any("err", err))
			return api.Logout401JSONResponse{
				UnauthorizedJSONResponse: Unauthorized(err),
			}, nil
		default:
			logger.Error("failed to logout", slog.Any("err", err))
			return api.Logout500JSONResponse{
				InternalErrorJSONResponse: InternalError(err),
			}, nil
		}
	}

	logger.Info("handler: logout success")

	return api.Logout200Response{}, nil
}
