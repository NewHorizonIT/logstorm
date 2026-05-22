package auth

import (
	"context"
	"time"
)

// Define the domain Account
type Account struct {
	ID        int
	Email     string
	Password  string
	CreatedAt time.Time
	UpdatedAt time.Time
}

// Define UseCase for service Auth
type AuthUsecase interface {
	Login(ctx context.Context, email, password string) (string, error)
	Register(ctx context.Context, email, password string) (string, error)
}

type AuthRepository interface {
	// Define methods for authentication repository here
	GetAccountByEmail(ctx context.Context, email string) (*Account, error)
	CreateAccount(ctx context.Context, account *Account) error
	GetAccountByID(ctx context.Context, id int) (*Account, error)
}
