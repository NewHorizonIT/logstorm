package auth

import (
	"context"
	"errors"

	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	authRepo AuthRepository
}

func NewAuthUsecase(repo AuthRepository) AuthUsecase {
	return &authUsecase{
		authRepo: repo,
	}
}

// Login implements [AuthUsecase].
func (a *authUsecase) Login(ctx context.Context, email string, password string) (string, error) {
	return "", errors.New("not implemented")
}

// Register implements [AuthUsecase].
func (a *authUsecase) Register(ctx context.Context, email string, password string) error {
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return err
	}

	user := &Account{
		Email:    email,
		Password: string(hash),
	}

	return a.authRepo.CreateAccount(ctx, user)
}