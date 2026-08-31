package auth

import (
	"context"
	"time"

	"github.com/google/uuid"
)

type CreateRefreshTokenParams struct {
	UserID    uuid.UUID
	TokenHash string
	ExpiresAt time.Time
}

type AuthRepository interface {
	CreateRefreshToken(ctx context.Context, params CreateRefreshTokenParams) (*RefreshToken, error)
	GetRefreshTokenByHash(ctx context.Context, tokenHash string) (*RefreshToken, error)
	RevokeRefreshToken(ctx context.Context, tokenHash string) error
	// RotateRefreshToken atomically revokes oldHash and creates a new token.
	// Prevents the window where revoke succeeds but create fails, leaving the user sessionless.
	RotateRefreshToken(ctx context.Context, oldHash string, newParams CreateRefreshTokenParams) (*RefreshToken, error)
	DeleteExpiredTokens(ctx context.Context) error
}
