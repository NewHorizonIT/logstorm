package user

import (
	"context"
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeUserRepository is a hand-written test double for UserRepository.
type fakeUserRepository struct {
	createFn     func(ctx context.Context, params CreateUserParams) (*User, error)
	getByIDFn    func(ctx context.Context, id uuid.UUID) (*User, error)
	getByEmailFn func(ctx context.Context, email string) (*User, error)
}

func (f *fakeUserRepository) Create(ctx context.Context, params CreateUserParams) (*User, error) {
	return f.createFn(ctx, params)
}

func (f *fakeUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*User, error) {
	return f.getByIDFn(ctx, id)
}

func (f *fakeUserRepository) GetByEmail(ctx context.Context, email string) (*User, error) {
	return f.getByEmailFn(ctx, email)
}

func fixedUser(email, fullName string) *User {
	return &User{
		ID:        uuid.New(),
		Email:     email,
		FullName:  fullName,
		Status:    "active",
		CreatedAt: time.Now(),
		UpdatedAt: time.Now(),
	}
}

func TestCreateUser_Success(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{
		createFn: func(_ context.Context, params CreateUserParams) (*User, error) {
			return fixedUser(params.Email, params.FullName), nil
		},
	}
	svc := NewUserService(repo)

	got, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "quan@example.com",
		PasswordHash: "hashed",
		FullName:     "Anh Quan",
	})

	require.NoError(t, err)
	assert.Equal(t, "quan@example.com", got.Email)
	assert.Equal(t, "Anh Quan", got.FullName)
}

func TestCreateUser_TrimsWhitespace(t *testing.T) {
	t.Parallel()

	var captured CreateUserParams
	repo := &fakeUserRepository{
		createFn: func(_ context.Context, params CreateUserParams) (*User, error) {
			captured = params
			return fixedUser(params.Email, params.FullName), nil
		},
	}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "  quan@example.com  ",
		PasswordHash: "hashed",
		FullName:     "  Anh Quan  ",
	})

	require.NoError(t, err)
	assert.Equal(t, "quan@example.com", captured.Email)
	assert.Equal(t, "Anh Quan", captured.FullName)
}

func TestCreateUser_EmailAlreadyExists(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{
		createFn: func(_ context.Context, _ CreateUserParams) (*User, error) {
			return nil, ErrUserEmailAlreadyExists
		},
	}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "quan@example.com",
		PasswordHash: "hashed",
		FullName:     "Anh Quan",
	})

	assert.ErrorIs(t, err, ErrUserEmailAlreadyExists)
}

func TestCreateUser_InvalidEmail(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "not-an-email",
		PasswordHash: "hashed",
		FullName:     "Anh Quan",
	})

	assert.ErrorIs(t, err, ErrValidation)
	assert.False(t, errors.Is(err, ErrUserEmailAlreadyExists), "should not reach the repository")
}

func TestCreateUser_EmptyEmail(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "   ",
		PasswordHash: "hashed",
		FullName:     "Anh Quan",
	})

	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateUser_EmptyFullName(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "quan@example.com",
		PasswordHash: "hashed",
		FullName:     "  ",
	})

	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateUser_FullNameTooLong(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "quan@example.com",
		PasswordHash: "hashed",
		FullName:     strings.Repeat("a", 101),
	})

	assert.ErrorIs(t, err, ErrValidation)
}

func TestCreateUser_FullNameAtMaxLength(t *testing.T) {
	t.Parallel()

	repo := &fakeUserRepository{
		createFn: func(_ context.Context, params CreateUserParams) (*User, error) {
			return fixedUser(params.Email, params.FullName), nil
		},
	}
	svc := NewUserService(repo)

	_, err := svc.CreateUser(context.Background(), CreateUserInput{
		Email:        "quan@example.com",
		PasswordHash: "hashed",
		FullName:     strings.Repeat("a", 100),
	})

	require.NoError(t, err)
}
