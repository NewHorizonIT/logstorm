package postgres

import (
	"context"
	"errors"

	"github.com/google/uuid"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/logstorm/api/internal/db"
	"github.com/logstorm/api/internal/modules/apikey"
)

type PostgresAPIKeyRepository struct {
	q *db.Queries
}

func NewPostgresAPIKeyRepository(pool *pgxpool.Pool) *PostgresAPIKeyRepository {
	return &PostgresAPIKeyRepository{q: db.New(pool)}
}

func (r *PostgresAPIKeyRepository) Create(ctx context.Context, params apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
	row, err := r.q.CreateAPIKey(ctx, db.CreateAPIKeyParams{
		ProjectID: pgtype.UUID{Bytes: params.ProjectID, Valid: true},
		OwnerID:   pgtype.UUID{Bytes: params.OwnerID, Valid: true},
		Name:      params.Name,
		KeyHash:   params.KeyHash,
		KeyPrefix: params.KeyPrefix,
	})
	if err != nil {
		return nil, err
	}
	return toAPIKey(row), nil
}

func (r *PostgresAPIKeyRepository) GetByHash(ctx context.Context, keyHash string) (*apikey.APIKey, error) {
	row, err := r.q.GetAPIKeyByHash(ctx, keyHash)
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apikey.ErrKeyNotFound
		}
		return nil, err
	}
	return toAPIKey(row), nil
}

func (r *PostgresAPIKeyRepository) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*apikey.APIKey, error) {
	rows, err := r.q.ListAPIKeysByOwner(ctx, pgtype.UUID{Bytes: ownerID, Valid: true})
	if err != nil {
		return nil, err
	}
	keys := make([]*apikey.APIKey, len(rows))
	for i, row := range rows {
		keys[i] = toAPIKey(row)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepository) ListByProject(ctx context.Context, projectID, ownerID uuid.UUID) ([]*apikey.APIKey, error) {
	rows, err := r.q.ListAPIKeysByProject(ctx, db.ListAPIKeysByProjectParams{
		ProjectID: pgtype.UUID{Bytes: projectID, Valid: true},
		OwnerID:   pgtype.UUID{Bytes: ownerID, Valid: true},
	})
	if err != nil {
		return nil, err
	}
	keys := make([]*apikey.APIKey, len(rows))
	for i, row := range rows {
		keys[i] = toAPIKey(row)
	}
	return keys, nil
}

func (r *PostgresAPIKeyRepository) Revoke(ctx context.Context, id, ownerID uuid.UUID) (*apikey.APIKey, error) {
	row, err := r.q.RevokeAPIKey(ctx, db.RevokeAPIKeyParams{
		ID:      pgtype.UUID{Bytes: id, Valid: true},
		OwnerID: pgtype.UUID{Bytes: ownerID, Valid: true},
	})
	if err != nil {
		if errors.Is(err, pgx.ErrNoRows) {
			return nil, apikey.ErrKeyNotFound
		}
		return nil, err
	}
	return toAPIKey(row), nil
}

func toAPIKey(row db.ApiKey) *apikey.APIKey {
	k := &apikey.APIKey{
		ID:        uuid.UUID(row.ID.Bytes),
		ProjectID: uuid.UUID(row.ProjectID.Bytes),
		OwnerID:   uuid.UUID(row.OwnerID.Bytes),
		Name:      row.Name,
		KeyHash:   row.KeyHash,
		KeyPrefix: row.KeyPrefix,
		CreatedAt: row.CreatedAt.Time,
	}
	if row.RevokedAt.Valid {
		t := row.RevokedAt.Time
		k.RevokedAt = &t
	}
	return k
}
