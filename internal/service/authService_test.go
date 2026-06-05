package service_test

import (
	"context"
	"errors"
	"testing"
	"time"

	masterconfig "github.com/LeHuuHai/server-management/config/master"
	authdomain "github.com/LeHuuHai/server-management/internal/domain/auth"
	apperr "github.com/LeHuuHai/server-management/internal/error"
	jwtprovider "github.com/LeHuuHai/server-management/internal/infra/jwt"
	"github.com/LeHuuHai/server-management/internal/model"
	"github.com/LeHuuHai/server-management/internal/service"
	"golang.org/x/crypto/bcrypt"
)

// --- mock AccountRepo ---

type mockAccountRepo struct {
	findByUserName func(userName string) (*model.Account, error)
	findByUserID   func(userID uint) (*model.Account, error)
}

func (m *mockAccountRepo) FindByUserName(userName string) (*model.Account, error) {
	return m.findByUserName(userName)
}

func (m *mockAccountRepo) FindByUserID(userID uint) (*model.Account, error) {
	return m.findByUserID(userID)
}

// --- mock TokenBlocklist ---

type mockBlocklist struct {
	revoke    func(ctx context.Context, token string, expiry time.Time) error
	isRevoked func(ctx context.Context, token string) (bool, error)
}

func (m *mockBlocklist) Revoke(ctx context.Context, token string, expiry time.Time) error {
	return m.revoke(ctx, token, expiry)
}

func (m *mockBlocklist) IsRevoked(ctx context.Context, token string) (bool, error) {
	return m.isRevoked(ctx, token)
}

// --- helpers ---

func newTestJWTProvider() *jwtprovider.JWTProvider {
	return jwtprovider.NewJWTProvider(&masterconfig.JWTConfig{
		AccessSecret:   "access-secret",
		RefreshSecret:  "refresh-secret",
		AccessExpired:  3600,
		RefreshExpired: 86400,
	})
}

func hashedPassword(t *testing.T, plain string) string {
	t.Helper()
	h, err := bcrypt.GenerateFromPassword([]byte(plain), bcrypt.MinCost)
	if err != nil {
		t.Fatalf("bcrypt: %v", err)
	}
	return string(h)
}

func newTestAccount(t *testing.T, password string) *model.Account {
	t.Helper()
	return &model.Account{
		ID:       1,
		UserID:   10,
		Username: "hai",
		Password: hashedPassword(t, password),
		Role:     authdomain.Role("admin"),
	}
}

// --- Login ---

func TestAuthService_Login_Success(t *testing.T) {
	account := newTestAccount(t, "secret")
	repo := &mockAccountRepo{
		findByUserName: func(_ string) (*model.Account, error) { return account, nil },
	}

	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, repo)
	result, err := svc.Login("hai", "secret")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.AccessToken == "" {
		t.Error("expected non-empty access token")
	}
	if result.RefreshToken == "" {
		t.Error("expected non-empty refresh token")
	}
}

func TestAuthService_Login_UserNotFound(t *testing.T) {
	repo := &mockAccountRepo{
		findByUserName: func(_ string) (*model.Account, error) {
			return nil, apperr.ErrRecordNotFound
		},
	}

	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, repo)
	_, err := svc.Login("notexist", "password")

	if !errors.Is(err, apperr.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	account := newTestAccount(t, "secret")
	repo := &mockAccountRepo{
		findByUserName: func(_ string) (*model.Account, error) { return account, nil },
	}

	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, repo)
	_, err := svc.Login("hai", "wrongpassword")

	if !errors.Is(err, apperr.ErrInvalidCredentials) {
		t.Errorf("expected ErrInvalidCredentials, got %v", err)
	}
}

func TestAuthService_Login_RepoError(t *testing.T) {
	repo := &mockAccountRepo{
		findByUserName: func(_ string) (*model.Account, error) {
			return nil, errors.New("db connection failed")
		},
	}

	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, repo)
	_, err := svc.Login("hai", "secret")

	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

// --- RefreshAccessToken ---

func TestAuthService_RefreshAccessToken_Success(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, err := jwt.GenerateRefreshToken(*account)
	if err != nil {
		t.Fatalf("generate refresh token: %v", err)
	}

	repo := &mockAccountRepo{
		findByUserID: func(_ uint) (*model.Account, error) { return account, nil },
	}
	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := service.NewAuthService(jwt, blocklist, repo)
	accessToken, err := svc.RefreshAccessToken(context.Background(), refreshToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if accessToken == "" {
		t.Error("expected non-empty access token")
	}
}

func TestAuthService_RefreshAccessToken_RevokedToken(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) { return true, nil },
	}

	svc := service.NewAuthService(jwt, blocklist, &mockAccountRepo{})
	_, err := svc.RefreshAccessToken(context.Background(), refreshToken)

	if !errors.Is(err, apperr.ErrRevokedToken) {
		t.Errorf("expected ErrRevokedToken, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_InvalidToken(t *testing.T) {
	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := service.NewAuthService(newTestJWTProvider(), blocklist, &mockAccountRepo{})
	_, err := svc.RefreshAccessToken(context.Background(), "invalid.token.string")

	if !errors.Is(err, apperr.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_BlocklistError(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) {
			return false, errors.New("redis down")
		},
	}

	svc := service.NewAuthService(jwt, blocklist, &mockAccountRepo{})
	_, err := svc.RefreshAccessToken(context.Background(), refreshToken)

	if err == nil {
		t.Error("expected error from blocklist, got nil")
	}
}

func TestAuthService_RefreshAccessToken_UserNotFound(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	repo := &mockAccountRepo{
		findByUserID: func(_ uint) (*model.Account, error) {
			return nil, apperr.ErrRecordNotFound
		},
	}
	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := service.NewAuthService(jwt, blocklist, repo)
	_, err := svc.RefreshAccessToken(context.Background(), refreshToken)

	if !errors.Is(err, apperr.ErrRecordNotFound) {
		t.Errorf("expected ErrRecordNotFound, got %v", err)
	}
}

func TestAuthService_RefreshAccessToken_RepoError(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	repo := &mockAccountRepo{
		findByUserID: func(_ uint) (*model.Account, error) {
			return nil, errors.New("db connection failed")
		},
	}
	blocklist := &mockBlocklist{
		isRevoked: func(_ context.Context, _ string) (bool, error) { return false, nil },
	}

	svc := service.NewAuthService(jwt, blocklist, repo)
	_, err := svc.RefreshAccessToken(context.Background(), refreshToken)

	if err == nil {
		t.Error("expected error from repo, got nil")
	}
}

// --- Logout ---

func TestAuthService_Logout_Success(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	var revokedToken string
	blocklist := &mockBlocklist{
		revoke: func(_ context.Context, token string, _ time.Time) error {
			revokedToken = token
			return nil
		},
	}

	svc := service.NewAuthService(jwt, blocklist, &mockAccountRepo{})
	err := svc.Logout(context.Background(), refreshToken)

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if revokedToken != refreshToken {
		t.Errorf("expected token %q to be revoked, got %q", refreshToken, revokedToken)
	}
}

func TestAuthService_Logout_InvalidToken(t *testing.T) {
	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, &mockAccountRepo{})
	err := svc.Logout(context.Background(), "bad.token")

	if !errors.Is(err, apperr.ErrInvalidToken) {
		t.Errorf("expected ErrInvalidToken, got %v", err)
	}
}

func TestAuthService_Logout_BlocklistError(t *testing.T) {
	account := newTestAccount(t, "secret")
	jwt := newTestJWTProvider()
	refreshToken, _ := jwt.GenerateRefreshToken(*account)

	blocklist := &mockBlocklist{
		revoke: func(_ context.Context, _ string, _ time.Time) error {
			return errors.New("redis down")
		},
	}

	svc := service.NewAuthService(jwt, blocklist, &mockAccountRepo{})
	err := svc.Logout(context.Background(), refreshToken)

	if err == nil {
		t.Error("expected error from blocklist, got nil")
	}
}

// --- HashPassword ---

func TestAuthService_HashPassword_Success(t *testing.T) {
	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, &mockAccountRepo{})
	hash, err := svc.HashPassword("mypassword")

	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if err := bcrypt.CompareHashAndPassword([]byte(hash), []byte("mypassword")); err != nil {
		t.Error("hash does not match original password")
	}
}

func TestAuthService_HashPassword_DifferentEachTime(t *testing.T) {
	svc := service.NewAuthService(newTestJWTProvider(), &mockBlocklist{}, &mockAccountRepo{})
	h1, _ := svc.HashPassword("same")
	h2, _ := svc.HashPassword("same")

	if h1 == h2 {
		t.Error("expected different hashes for same password due to bcrypt salt")
	}
}
