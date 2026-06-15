package auth

import (
	"context"
	"time"
)

// Define the domain Account
type Account struct {
	ID           int
	Email        string
	PasswordHash string
	CreatedAt    time.Time
	UpdatedAt    time.Time
}

// Define UseCase for service Auth
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) (*RegisterResult, error)
	Refresh(ctx context.Context, refreshToken string) (*HandleRefreshTokenResult, error)
	Logout(ctx context.Context, refreshToken string) error
}

type AccountRepository interface {
	// Define methods for authentication repository here
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id int) (*Account, error)
}

type SessionRepository interface {
	StoreRefreshToken(ctx context.Context, sid string, userID int, expiration time.Duration) error
	DeleteSession(ctx context.Context, sid string) error
	GetUserIDBySID(ctx context.Context, sid string) (int, error)
}
