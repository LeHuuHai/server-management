package handler_test

import (
	"context"
	"errors"
	"testing"

	"github.com/LeHuuHai/server-management/api"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	"github.com/LeHuuHai/server-management/internal/handler"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/mocks"
	"github.com/stretchr/testify/assert"
)

func TestAuthHandler_Login_Success(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Login("hai", "secret").Return(&model.LoginResult{
		AccessToken:  "access-tok",
		RefreshToken: "refresh-tok",
	}, nil)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Login(context.Background(), api.LoginRequestObject{
		Body: &api.LoginJSONRequestBody{Username: "hai", Password: "secret"},
	})

	assert.NoError(t, err)
	res, ok := resp.(api.Login200JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, "access-tok", *res.AccessToken)
	assert.Equal(t, "refresh-tok", *res.RefreshToken)
}

func TestAuthHandler_Login_UserNotFound_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Login("nobody", "pass").Return(nil, apperr.ErrRecordNotFound)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Login(context.Background(), api.LoginRequestObject{
		Body: &api.LoginJSONRequestBody{Username: "nobody", Password: "pass"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Login401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_Login_InvalidCredentials_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Login("hai", "wrongpass").Return(nil, apperr.ErrInvalidCredentials)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Login(context.Background(), api.LoginRequestObject{
		Body: &api.LoginJSONRequestBody{Username: "hai", Password: "wrongpass"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Login401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_Login_SignTokenError_Returns500(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Login("hai", "secret").Return(nil, apperr.ErrSignToken)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Login(context.Background(), api.LoginRequestObject{
		Body: &api.LoginJSONRequestBody{Username: "hai", Password: "secret"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Login500JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_Login_UnexpectedError_Returns500(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Login("hai", "secret").Return(nil, errors.New("db down"))

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Login(context.Background(), api.LoginRequestObject{
		Body: &api.LoginJSONRequestBody{Username: "hai", Password: "secret"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Login500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// RefreshToken
// ---------------------------------------------------------------------------

func TestAuthHandler_RefreshToken_Success(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "valid-refresh").Return("new-access", nil)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "valid-refresh"},
	})

	assert.NoError(t, err)
	res, ok := resp.(api.RefreshToken200JSONResponse)
	assert.True(t, ok)
	assert.Equal(t, "new-access", *res.AccessToken)
	assert.Equal(t, "valid-refresh", *res.RefreshToken)
}

func TestAuthHandler_RefreshToken_InvalidToken_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "bad-token").Return("", apperr.ErrInvalidToken)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "bad-token"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.RefreshToken401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_RefreshToken_RevokedToken_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "revoked").Return("", apperr.ErrRevokedToken)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "revoked"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.RefreshToken401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_RefreshToken_UserNotFound_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "unknow-token").Return("", apperr.ErrRecordNotFound)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "unknow-token"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.RefreshToken401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_RefreshToken_SignTokenError_Returns500(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "tok").Return("", apperr.ErrSignToken)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "tok"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.RefreshToken500JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_RefreshToken_UnexpectedError_Returns500(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().RefreshAccessToken(context.Background(), "tok").Return("", errors.New("redis down"))

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.RefreshToken(context.Background(), api.RefreshTokenRequestObject{
		Body: &api.RefreshTokenJSONRequestBody{RefreshToken: "tok"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.RefreshToken500JSONResponse)
	assert.True(t, ok)
}

// ---------------------------------------------------------------------------
// Logout
// ---------------------------------------------------------------------------

func TestAuthHandler_Logout_Success(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Logout(context.Background(), "valid-refresh").Return(nil)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Logout(context.Background(), api.LogoutRequestObject{
		Body: &api.LogoutJSONRequestBody{RefreshToken: "valid-refresh"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Logout200Response)
	assert.True(t, ok)
}

func TestAuthHandler_Logout_InvalidToken_Returns401(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Logout(context.Background(), "bad").Return(apperr.ErrInvalidToken)

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Logout(context.Background(), api.LogoutRequestObject{
		Body: &api.LogoutJSONRequestBody{RefreshToken: "bad"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Logout401JSONResponse)
	assert.True(t, ok)
}

func TestAuthHandler_Logout_UnexpectedError_Returns500(t *testing.T) {
	mockAuth := mocks.NewMockAuthServiceInterface(t)
	mockAuth.EXPECT().Logout(context.Background(), "tok").Return(errors.New("redis down"))

	handler := handler.NewAuthHandler(mockAuth)
	resp, err := handler.Logout(context.Background(), api.LogoutRequestObject{
		Body: &api.LogoutJSONRequestBody{RefreshToken: "tok"},
	})

	assert.NoError(t, err)
	_, ok := resp.(api.Logout500JSONResponse)
	assert.True(t, ok)
}
