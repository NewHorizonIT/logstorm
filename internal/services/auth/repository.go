package auth

import (
	"context"

	"gorm.io/gorm"
)

// Define repository layer for authentication service here
type authRepository struct {
	db *gorm.DB
}

// Initialize authRepository
func NewAuthRepository(db *gorm.DB) AuthRepository {
	return &authRepository{
		db: db,
	}
}

// CreateAccount implements [AuthRepository].
func (a *authRepository) CreateAccount(ctx context.Context, account *Account) error {
	panic("unimplemented")
}

// GetAccountByEmail implements [AuthRepository].
func (a *authRepository) GetAccountByEmail(ctx context.Context, email string) (*Account, error) {
	panic("unimplemented")
}

// GetAccountByID implements [AuthRepository].
func (a *authRepository) GetAccountByID(ctx context.Context, id string) (*Account, error) {
	panic("unimplemented")
}
