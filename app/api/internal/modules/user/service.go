package user

import (
	"context"
	"fmt"
	"net/mail"
	"strings"
)

type CreateUserInput struct {
	Email        string
	PasswordHash string
	FullName     string
}

type UserService struct {
	repo UserRepository
}

func NewUserService(repo UserRepository) *UserService {
	return &UserService{repo: repo}
}

func (s *UserService) CreateUser(ctx context.Context, input CreateUserInput) (*User, error) {
	input.Email = strings.TrimSpace(input.Email)
	input.FullName = strings.TrimSpace(input.FullName)

	if input.PasswordHash == "" {
		return nil, fmt.Errorf("%w: password hash is required", ErrValidation)
	}
	if input.Email == "" {
		return nil, fmt.Errorf("%w: email is required", ErrValidation)
	}
	if _, err := mail.ParseAddress(input.Email); err != nil {
		return nil, fmt.Errorf("%w: email is invalid", ErrValidation)
	}
	if input.FullName == "" {
		return nil, fmt.Errorf("%w: full name is required", ErrValidation)
	}
	if len([]rune(input.FullName)) > 100 {
		return nil, fmt.Errorf("%w: full name must not exceed 100 characters", ErrValidation)
	}

	return s.repo.Create(ctx, CreateUserParams{
		Email:        input.Email,
		PasswordHash: input.PasswordHash,
		FullName:     input.FullName,
	})
}
