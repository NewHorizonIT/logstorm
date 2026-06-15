package auth

import (
	"context"
	"errors"
	"time"

	"github.com/NewHorizonIT/logstorm/internal/infra/redis"
	"github.com/NewHorizonIT/logstorm/pkg"
	"golang.org/x/crypto/bcrypt"
)

type authUsecase struct {
	authRepo    AccountRepository
	sessionRepo SessionRepository
}

func NewAuthUsecase(repo AccountRepository, sessionRepo SessionRepository) AuthUsecase {
	return &authUsecase{
		authRepo:    repo,
		sessionRepo: sessionRepo,
	}
}

// Login implements [AuthUsecase].
func (a *authUsecase) Login(ctx context.Context, email string, password string) (string, error) {
	account, err := a.authRepo.GetAccountByEmail(ctx, email)
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	if account == nil {
		return "", errors.New("invalid credentials")
	}

	err = bcrypt.CompareHashAndPassword([]byte(account.PasswordHash), []byte(password))
	if err != nil {
		return "", errors.New("invalid credentials")
	}

	// Generate JWT token
	token, err := pkg.GenerateToken(account.ID, 15*time.Minute)
	if err != nil {
		return "", err
	}

	return token, nil
}

// Register implements [AuthUsecase].
func (a *authUsecase) Register(ctx context.Context, email string, password string) (*RegisterResult, error) {
	account, err := a.authRepo.GetAccountByEmail(ctx, email)

	if err != nil {
		return nil, err
	}

	if account != nil {
		return nil, errors.New("email already in use")
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, errors.New("failed to hash password")
	}

	user := &Account{
		Email:        email,
		PasswordHash: string(hash),
	}

	err = a.authRepo.CreateAccount(ctx, user)
	if err != nil {
		return nil, errors.New("failed to create account")
	}

	// generate token for the new user
	accessToken, err := pkg.GenerateToken(user.ID, 15*time.Minute) // Token valid for 15 minutes
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	refreshToken, err := pkg.GenerateToken(user.ID, 7*24*time.Hour) // Refresh token valid for 7 days
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &RegisterResult{
		Account:      user,
		AccessToken:  accessToken,
		RefreshToken: refreshToken,
	}, nil
}

// Refresh implements [AuthUsecase].
func (a *authUsecase) Refresh(ctx context.Context, refreshToken string) (*HandleRefreshTokenResult, error) {
	// Step 1: Verify the refresh token and get the user ID
	payload, err := pkg.VerifyToken(refreshToken)
	if err != nil {
		return nil, errors.New("invalid refresh token")
	}

	// Step 2: Get refresh token from Redis and compare with the provided token (implement in future)

	// Step 3:Generate new access token
	accessToken, err := pkg.GenerateToken(payload.UserID, 15*time.Minute)
	if err != nil {
		return nil, errors.New("failed to generate access token")
	}

	// Step 4: Generate new refresh token
	newRefreshToken, err := pkg.GenerateToken(payload.UserID, 7*24*time.Hour)
	if err != nil {
		return nil, errors.New("failed to generate refresh token")
	}

	return &HandleRefreshTokenResult{
		AccessToken:  accessToken,
		RefreshToken: newRefreshToken,
	}, nil
}

// Logout implements [AuthUsecase].
func (a *authUsecase) Logout(ctx context.Context, refreshToken string) error {
	// Step 1: Verify the refresh token and get the user ID
	payload, err := pkg.VerifyToken(refreshToken)
	if err != nil {
		return errors.New("invalid refresh token")
	}

	// Step 2: Delete the refresh token from Redis (implement in future)
	redisKey := redis.SessionKey(payload.Sid)

	err = a.sessionRepo.DeleteSession(ctx, redisKey)
	if err != nil {
		return errors.New("failed to invalidate refresh token")
	}

	return nil
}
