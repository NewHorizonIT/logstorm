package auth_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"golang.org/x/crypto/bcrypt"

	"github.com/NewHorizonIT/logstorm/internal/global"
	auth "github.com/NewHorizonIT/logstorm/internal/services/auth"
	"github.com/NewHorizonIT/logstorm/pkg"
)

type mockRepo struct {
	acct        *auth.Account
	getErr      error
	createErr   error
	createdWith *auth.Account
}

func (m *mockRepo) GetAccountByEmail(ctx context.Context, email string) (*auth.Account, error) {
	if m.getErr != nil {
		return nil, m.getErr
	}
	return m.acct, nil
}

func (m *mockRepo) CreateAccount(ctx context.Context, account *auth.Account) error {
	if m.createErr != nil {
		return m.createErr
	}
	// simulate DB assigning ID
	account.ID = 7
	m.createdWith = account
	return nil
}

func (m *mockRepo) GetAccountByID(ctx context.Context, id int) (*auth.Account, error) {
	if m.acct != nil && m.acct.ID == id {
		return m.acct, nil
	}
	return nil, errors.New("not found")
}

func TestAuthUsecase_Login(t *testing.T) {
	global.GlobalConfig.JWTConfig.Secret = "test-secret"

	t.Run("success", func(t *testing.T) {
		password := "s3cretpw"

		hash, _ := bcrypt.GenerateFromPassword(
			[]byte(password),
			bcrypt.DefaultCost,
		)

		repo := &mockRepo{
			acct: &auth.Account{
				ID:           123,
				Email:        "u@e.com",
				PasswordHash: string(hash),
			},
		}

		u := auth.NewAuthUsecase(repo)

		token, err := u.Login(
			context.Background(),
			"u@e.com",
			password,
		)

		if err != nil {
			t.Fatalf("expected nil error, got %v", err)
		}

		if token == "" {
			t.Fatal("expected token")
		}
	})

	t.Run("wrong_password", func(t *testing.T) {
		hash, _ := bcrypt.GenerateFromPassword(
			[]byte("correct-password"),
			bcrypt.DefaultCost,
		)

		repo := &mockRepo{
			acct: &auth.Account{
				ID:           123,
				Email:        "u@e.com",
				PasswordHash: string(hash),
			},
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Login(
			context.Background(),
			"u@e.com",
			"wrong-password",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("account_not_found", func(t *testing.T) {
		repo := &mockRepo{
			acct: nil,
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Login(
			context.Background(),
			"missing@e.com",
			"password",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("repository_error", func(t *testing.T) {
		repo := &mockRepo{
			getErr: errors.New("database down"),
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Login(
			context.Background(),
			"u@e.com",
			"password",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAuthUsecase_Register(t *testing.T) {
	global.GlobalConfig.JWTConfig.Secret = "test-secret"

	t.Run("email_already_exists", func(t *testing.T) {
		repo := &mockRepo{
			acct: &auth.Account{
				ID:    1,
				Email: "taken@e.com",
			},
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Register(
			context.Background(),
			"taken@e.com",
			"password123",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("success", func(t *testing.T) {
		repo := &mockRepo{}

		u := auth.NewAuthUsecase(repo)

		res, err := u.Register(
			context.Background(),
			"new@e.com",
			"password123",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res == nil {
			t.Fatal("expected result")
		}

		if res.Account == nil {
			t.Fatal("expected account")
		}

		if res.Account.ID != 7 {
			t.Fatalf(
				"expected ID=7 got %d",
				res.Account.ID,
			)
		}

		if res.AccessToken == "" {
			t.Fatal("expected access token")
		}

		if res.RefreshToken == "" {
			t.Fatal("expected refresh token")
		}
	})

	t.Run("password_is_hashed", func(t *testing.T) {
		repo := &mockRepo{}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Register(
			context.Background(),
			"user@e.com",
			"password123",
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if repo.createdWith == nil {
			t.Fatal("expected account passed to repository")
		}

		if repo.createdWith.PasswordHash == "password123" {
			t.Fatal("password stored in plain text")
		}

		err = bcrypt.CompareHashAndPassword(
			[]byte(repo.createdWith.PasswordHash),
			[]byte("password123"),
		)

		if err != nil {
			t.Fatal("password hash invalid")
		}
	})

	t.Run("repository_error_on_lookup", func(t *testing.T) {
		repo := &mockRepo{
			getErr: errors.New("database down"),
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Register(
			context.Background(),
			"user@e.com",
			"password123",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("repository_error_on_create", func(t *testing.T) {
		repo := &mockRepo{
			createErr: errors.New("insert failed"),
		}

		u := auth.NewAuthUsecase(repo)

		_, err := u.Register(
			context.Background(),
			"user@e.com",
			"password123",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}

func TestAuthUsecase_Refresh(t *testing.T) {
	global.GlobalConfig.JWTConfig.Secret = "test-secret"

	t.Run("success", func(t *testing.T) {
		u := auth.NewAuthUsecase(nil)

		refreshToken, err := pkg.GenerateToken(
			123,
			7*24*time.Hour,
		)

		if err != nil {
			t.Fatalf("failed to create test token: %v", err)
		}

		res, err := u.Refresh(
			context.Background(),
			refreshToken,
		)

		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}

		if res == nil {
			t.Fatal("expected result")
		}

		if res.AccessToken == "" {
			t.Fatal("expected access token")
		}

		if res.RefreshToken == "" {
			t.Fatal("expected refresh token")
		}

		if res.AccessToken == refreshToken {
			t.Fatal("access token should be newly generated")
		}
	})

	t.Run("invalid_token", func(t *testing.T) {
		u := auth.NewAuthUsecase(nil)

		_, err := u.Refresh(
			context.Background(),
			"this-is-not-a-jwt",
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})

	t.Run("expired_token", func(t *testing.T) {
		u := auth.NewAuthUsecase(nil)

		expiredToken, err := pkg.GenerateToken(
			123,
			-time.Minute,
		)

		if err != nil {
			t.Fatalf("failed to create token: %v", err)
		}

		_, err = u.Refresh(
			context.Background(),
			expiredToken,
		)

		if err == nil {
			t.Fatal("expected error")
		}
	})
}
