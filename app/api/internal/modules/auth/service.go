package auth

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"errors"
	"strings"
	"time"

	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
	"github.com/logstorm/api/internal/modules/user"
)

type RegisterInput struct {
	Email    string
	FullName string
	Password string
}

type LoginInput struct {
	Email    string
	Password string
}

type AuthService struct {
	userProvider UserProvider
	authRepo     AuthRepository
	tokenSvc     *TokenService
}

func NewAuthService(userProvider UserProvider, authRepo AuthRepository, tokenSvc *TokenService) *AuthService {
	return &AuthService{
		userProvider: userProvider,
		authRepo:     authRepo,
		tokenSvc:     tokenSvc,
	}
}

func (s *AuthService) Register(ctx context.Context, input RegisterInput) (*user.User, error) {
	passwordHash, err := bcrypt.GenerateFromPassword([]byte(input.Password), bcrypt.DefaultCost)
	if err != nil {
		return nil, err
	}
	return s.userProvider.CreateUser(ctx, user.CreateUserInput{
		Email:        strings.TrimSpace(input.Email),
		FullName:     strings.TrimSpace(input.FullName),
		PasswordHash: string(passwordHash),
	})
}

func (s *AuthService) Login(ctx context.Context, input LoginInput) (accessToken, rawRefreshToken string, err error) {
	u, err := s.userProvider.GetByEmail(ctx, strings.TrimSpace(input.Email))
	if err != nil {
		if errors.Is(err, user.ErrUserNotFound) {
			return "", "", ErrInvalidCredentials
		}
		return "", "", err
	}

	if err := bcrypt.CompareHashAndPassword([]byte(u.PasswordHash), []byte(input.Password)); err != nil {
		return "", "", ErrInvalidCredentials
	}

	return s.issueTokenPair(ctx, u.ID)
}

func (s *AuthService) Logout(ctx context.Context, rawRefreshToken string) error {
	return s.authRepo.RevokeRefreshToken(ctx, hashToken(rawRefreshToken))
}

func (s *AuthService) Refresh(ctx context.Context, rawRefreshToken string) (accessToken, newRawRefreshToken string, err error) {
	hashed := hashToken(rawRefreshToken)
	existing, err := s.authRepo.GetRefreshTokenByHash(ctx, hashed)
	if err != nil {
		if errors.Is(err, ErrRefreshTokenNotFound) {
			return "", "", ErrRefreshTokenNotFound
		}
		return "", "", err
	}

	accessToken, err = s.tokenSvc.GenerateAccessToken(existing.UserID)
	if err != nil {
		return "", "", err
	}

	var hashedRefresh string
	newRawRefreshToken, hashedRefresh, err = s.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if _, err = s.authRepo.RotateRefreshToken(ctx, hashed, CreateRefreshTokenParams{
		UserID:    existing.UserID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenSvc.RefreshTokenTTL()),
	}); err != nil {
		return "", "", err
	}

	return accessToken, newRawRefreshToken, nil
}

func (s *AuthService) issueTokenPair(ctx context.Context, userID uuid.UUID) (accessToken, rawRefreshToken string, err error) {
	accessToken, err = s.tokenSvc.GenerateAccessToken(userID)
	if err != nil {
		return "", "", err
	}

	var hashedRefresh string
	rawRefreshToken, hashedRefresh, err = s.tokenSvc.GenerateRefreshToken()
	if err != nil {
		return "", "", err
	}

	if _, err := s.authRepo.CreateRefreshToken(ctx, CreateRefreshTokenParams{
		UserID:    userID,
		TokenHash: hashedRefresh,
		ExpiresAt: time.Now().Add(s.tokenSvc.RefreshTokenTTL()),
	}); err != nil {
		return "", "", err
	}

	return accessToken, rawRefreshToken, nil
}

func hashToken(raw string) string {
	sum := sha256.Sum256([]byte(raw))
	return hex.EncodeToString(sum[:])
}
