package auth_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/crypto/bcrypt"
)

// --- fakes ---

type fakeUserProvider struct {
	createUserFn func(ctx context.Context, input user.CreateUserInput) (*user.User, error)
	getByEmailFn func(ctx context.Context, email string) (*user.User, error)
}

func (f *fakeUserProvider) CreateUser(ctx context.Context, input user.CreateUserInput) (*user.User, error) {
	return f.createUserFn(ctx, input)
}

func (f *fakeUserProvider) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	return f.getByEmailFn(ctx, email)
}

type fakeAuthRepo struct {
	createFn    func(ctx context.Context, p auth.CreateRefreshTokenParams) (*auth.RefreshToken, error)
	getByHashFn func(ctx context.Context, hash string) (*auth.RefreshToken, error)
	revokeFn    func(ctx context.Context, hash string) error
	deleteFn    func(ctx context.Context) error
}

func (f *fakeAuthRepo) CreateRefreshToken(ctx context.Context, p auth.CreateRefreshTokenParams) (*auth.RefreshToken, error) {
	return f.createFn(ctx, p)
}

func (f *fakeAuthRepo) GetRefreshTokenByHash(ctx context.Context, hash string) (*auth.RefreshToken, error) {
	return f.getByHashFn(ctx, hash)
}

func (f *fakeAuthRepo) RevokeRefreshToken(ctx context.Context, hash string) error {
	return f.revokeFn(ctx, hash)
}

func (f *fakeAuthRepo) DeleteExpiredTokens(ctx context.Context) error {
	if f.deleteFn != nil {
		return f.deleteFn(ctx)
	}
	return nil
}

func setupService(t *testing.T) (*auth.AuthService, *fakeUserProvider, *fakeAuthRepo) {
	t.Helper()
	userProvider := &fakeUserProvider{}
	authRepo := &fakeAuthRepo{}
	tokenSvc := auth.NewTokenService(config.AuthConfig{
		JWTSecret:       "test-secret-key-minimum-32-chars!!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
	svc := auth.NewAuthService(userProvider, authRepo, tokenSvc)
	return svc, userProvider, authRepo
}

func hashedPassword(t *testing.T, raw string) string {
	t.Helper()
	b, err := bcrypt.GenerateFromPassword([]byte(raw), bcrypt.MinCost)
	require.NoError(t, err)
	return string(b)
}

// --- Register ---

func TestAuthService_Register_Success(t *testing.T) {
	t.Parallel()

	svc, userProvider, _ := setupService(t)
	ctx := context.Background()

	want := &user.User{ID: uuid.New(), Email: "new@example.com", FullName: "New User"}
	userProvider.createUserFn = func(_ context.Context, input user.CreateUserInput) (*user.User, error) {
		assert.Equal(t, "new@example.com", input.Email)
		assert.Equal(t, "New User", input.FullName)
		assert.NotEmpty(t, input.PasswordHash)
		return want, nil
	}

	got, err := svc.Register(ctx, auth.RegisterInput{
		Email:    "new@example.com",
		FullName: "New User",
		Password: "securepassword",
	})

	require.NoError(t, err)
	assert.Equal(t, want.ID, got.ID)
}

func TestAuthService_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()

	svc, userProvider, _ := setupService(t)
	ctx := context.Background()

	userProvider.createUserFn = func(_ context.Context, _ user.CreateUserInput) (*user.User, error) {
		return nil, user.ErrUserEmailAlreadyExists
	}

	_, err := svc.Register(ctx, auth.RegisterInput{
		Email:    "dup@example.com",
		FullName: "Dup User",
		Password: "password123",
	})

	assert.ErrorIs(t, err, user.ErrUserEmailAlreadyExists)
}

// --- Login ---

func TestAuthService_Login_Success(t *testing.T) {
	t.Parallel()

	svc, userProvider, authRepo := setupService(t)
	ctx := context.Background()

	u := &user.User{
		ID:           uuid.New(),
		Email:        "user@example.com",
		PasswordHash: hashedPassword(t, "mypassword"),
	}
	userProvider.getByEmailFn = func(_ context.Context, _ string) (*user.User, error) {
		return u, nil
	}
	authRepo.createFn = func(_ context.Context, _ auth.CreateRefreshTokenParams) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{ID: uuid.New(), UserID: u.ID}, nil
	}

	accessToken, rawRefresh, err := svc.Login(ctx, auth.LoginInput{
		Email:    "user@example.com",
		Password: "mypassword",
	})

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, rawRefresh)
}

func TestAuthService_Login_WrongPassword(t *testing.T) {
	t.Parallel()

	svc, userProvider, _ := setupService(t)
	ctx := context.Background()

	u := &user.User{
		ID:           uuid.New(),
		PasswordHash: hashedPassword(t, "correct"),
	}
	userProvider.getByEmailFn = func(_ context.Context, _ string) (*user.User, error) {
		return u, nil
	}

	_, _, err := svc.Login(ctx, auth.LoginInput{Email: "x@x.com", Password: "wrong"})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

func TestAuthService_Login_EmailNotFound(t *testing.T) {
	t.Parallel()

	svc, userProvider, _ := setupService(t)
	ctx := context.Background()

	userProvider.getByEmailFn = func(_ context.Context, _ string) (*user.User, error) {
		return nil, user.ErrUserNotFound
	}

	_, _, err := svc.Login(ctx, auth.LoginInput{Email: "ghost@x.com", Password: "any"})

	assert.ErrorIs(t, err, auth.ErrInvalidCredentials)
}

// --- Logout ---

func TestAuthService_Logout(t *testing.T) {
	t.Parallel()

	svc, _, authRepo := setupService(t)
	ctx := context.Background()

	var revokedHash string
	authRepo.revokeFn = func(_ context.Context, hash string) error {
		revokedHash = hash
		return nil
	}

	err := svc.Logout(ctx, "raw-refresh-token")

	require.NoError(t, err)
	assert.NotEmpty(t, revokedHash)
	assert.NotEqual(t, "raw-refresh-token", revokedHash)
}

// --- Refresh ---

func TestAuthService_Refresh_Success(t *testing.T) {
	t.Parallel()

	svc, _, authRepo := setupService(t)
	ctx := context.Background()
	userID := uuid.New()

	authRepo.getByHashFn = func(_ context.Context, _ string) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{ID: uuid.New(), UserID: userID}, nil
	}
	authRepo.revokeFn = func(_ context.Context, _ string) error { return nil }
	authRepo.createFn = func(_ context.Context, _ auth.CreateRefreshTokenParams) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{ID: uuid.New(), UserID: userID}, nil
	}

	accessToken, newRawRefresh, err := svc.Refresh(ctx, "old-raw-token")

	require.NoError(t, err)
	assert.NotEmpty(t, accessToken)
	assert.NotEmpty(t, newRawRefresh)
}

func TestAuthService_Refresh_InvalidToken(t *testing.T) {
	t.Parallel()

	svc, _, authRepo := setupService(t)
	ctx := context.Background()

	authRepo.getByHashFn = func(_ context.Context, _ string) (*auth.RefreshToken, error) {
		return nil, auth.ErrRefreshTokenNotFound
	}

	_, _, err := svc.Refresh(ctx, "bogus-token")

	assert.ErrorIs(t, err, auth.ErrRefreshTokenNotFound)
}
