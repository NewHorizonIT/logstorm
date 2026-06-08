package auth

import (
	"context"
	"errors"

	"gorm.io/gorm"
)

// Define repository layer for authentication service here
type authRepository struct {
	db *gorm.DB
}

// Initialize authRepository
func NewAuthRepository(db *gorm.DB) AccountRepository {
	return &authRepository{
		db: db,
	}
}

// CreateAccount implements [AccountRepository].
func (a *authRepository) CreateAccount(ctx context.Context, account *Account) error {
	err := a.db.WithContext(ctx).Create(account).Error
	if err != nil {
		return err
	}
	return nil
}

// GetAccountByEmail implements [AccountRepository].
func (a *authRepository) GetAccountByEmail(
	ctx context.Context,
	email string,
) (*Account, error) {

	var account Account

	err := a.db.
		WithContext(ctx).
		Where("email = ?", email).
		First(&account).
		Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}

	if err != nil {
		return nil, err
	}

	return &account, nil
}

// GetAccountByID implements [AccountRepository].
func (a *authRepository) GetAccountByID(ctx context.Context, id int) (*Account, error) {
	var account Account
	err := a.db.WithContext(ctx).Where("id = ?", id).First(&account).Error
	if err != nil {
		return nil, err
	}
	return &account, nil
}
