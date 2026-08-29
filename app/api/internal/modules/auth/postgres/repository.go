package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/db"
	"github.com/logstorm/api/internal/modules/auth"
)

type PostgresAuthRepository struct {
	queries *db.Queries
}

func NewPostgresAuthRepository(pool *pgxpool.Pool) *PostgresAuthRepository {
	return &PostgresAuthRepository{queries: db.New(pool)}
}

func (r *PostgresAuthRepository) CreateRefreshToken(
	ctx context.Context,
	params auth.CreateRefreshTokenParams,
) (*auth.RefreshToken, error) {
	row, err := r.queries.CreateRefreshToken(ctx, db.CreateRefreshTokenParams{
		UserID:    pgtype.UUID{Bytes: params.UserID, Valid: true},
		TokenHash: params.TokenHash,
		ExpiresAt: pgtype.Timestamptz{Time: params.ExpiresAt, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	return toRefreshToken(row), nil
}

func (r *PostgresAuthRepository) GetRefreshTokenByHash(
	ctx context.Context,
	tokenHash string,
) (*auth.RefreshToken, error) {
	row, err := r.queries.GetRefreshTokenByHash(ctx, tokenHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, auth.ErrRefreshTokenNotFound
		}
		return nil, err
	}
	return toRefreshToken(row), nil
}

func (r *PostgresAuthRepository) RevokeRefreshToken(ctx context.Context, tokenHash string) error {
	return r.queries.RevokeRefreshToken(ctx, tokenHash)
}

func (r *PostgresAuthRepository) DeleteExpiredTokens(ctx context.Context) error {
	return r.queries.DeleteExpiredRefreshTokens(ctx)
}

func toRefreshToken(row db.RefreshToken) *auth.RefreshToken {
	t := &auth.RefreshToken{
		ID:        uuid.UUID(row.ID.Bytes),
		UserID:    uuid.UUID(row.UserID.Bytes),
		TokenHash: row.TokenHash,
		ExpiresAt: row.ExpiresAt.Time,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.RevokedAt.Valid {
		rv := row.RevokedAt.Time
		t.RevokedAt = &rv
	}
	return t
}
