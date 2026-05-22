package auth

import (
	"context"
	"errors"

	"github.com/NewHorizonIT/logstorm/pkg"
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
	account, err := a.authRepo.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.Password), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := pkg.GenerateToken(account.ID)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register implements [AuthUsecase].
func (a *authUsecase) Register(ctx context.Context, email string, password string) (string, error) {
	account, err := a.authRepo.GetAccountByEmail(ctx, email)
	if err == nil && account != nil {
		return "", errors.New("email already in use")
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}

	user := &Account{
		Email:    email,
		Password: string(hash),
	}

	err = a.authRepo.CreateAccount(ctx, user)
	if err != nil {
		return "", err
	}

	// generate token for the new user
	token, err := pkg.GenerateToken(user.ID)

	return token, nil
}
