package user

import (
	"context"

	"github.com/google/uuid"
)

type CreateUserParams struct {
	Email        string
	PasswordHash string
	FullName     string
}

type UserRepository interface {
	Create(ctx context.Context, params CreateUserParams) (*User, error)
	GetByID(ctx context.Context, id uuid.UUID) (*User, error)
	GetByEmail(ctx context.Context, email string) (*User, error)
}
