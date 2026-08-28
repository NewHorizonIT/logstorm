package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/db"
	"github.com/logstorm/api/internal/modules/user"
)

const uniqueViolationCode = "23505"

type PostgresUserRepository struct {
	queries *db.Queries
}

func NewPostgresUserRepository(pool *pgxpool.Pool) *PostgresUserRepository {
	return &PostgresUserRepository{queries: db.New(pool)}
}

func (r *PostgresUserRepository) Create(ctx context.Context, params user.CreateUserParams) (*user.User, error) {
	row, err := r.queries.CreateUser(ctx, db.CreateUserParams{
		Email:        params.Email,
		PasswordHash: params.PasswordHash,
		FullName:     params.FullName,
	})
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == uniqueViolationCode {
			return nil, user.ErrUserEmailAlreadyExists
		}
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *PostgresUserRepository) GetByID(ctx context.Context, id uuid.UUID) (*user.User, error) {
	row, err := r.queries.GetUserByID(ctx, pgtype.UUID{Bytes: id, Valid: true})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(row), nil
}

func (r *PostgresUserRepository) GetByEmail(ctx context.Context, email string) (*user.User, error) {
	row, err := r.queries.GetUserByEmail(ctx, email)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, user.ErrUserNotFound
		}
		return nil, err
	}
	return toDomainUser(row), nil
}

func toDomainUser(row db.User) *user.User {
	u := &user.User{
		ID:           uuid.UUID(row.ID.Bytes),
		Email:        row.Email,
		PasswordHash: row.PasswordHash,
		FullName:     row.FullName,
		AvatarURL:    row.AvatarUrl.String,
		Status:       row.Status,
		IsVerified:   row.IsVerified,
		CreatedAt:    row.CreatedAt.Time,
		UpdatedAt:    row.UpdatedAt.Time,
	}
	if row.LastLoginAt.Valid {
		t := row.LastLoginAt.Time
		u.LastLoginAt = &t
	}
	return u
}

