package apikey

import (
	"context"

	"github.com/google/uuid"
)

type CreateAPIKeyParams struct {
	ProjectID uuid.UUID
	OwnerID   uuid.UUID
	Name      string
	KeyHash   string
	KeyPrefix string
}

type APIKeyRepository interface {
	Create(ctx context.Context, params CreateAPIKeyParams) (*APIKey, error)
	GetByHash(ctx context.Context, keyHash string) (*APIKey, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*APIKey, error)
	ListByProject(ctx context.Context, projectID, ownerID uuid.UUID) ([]*APIKey, error)
	Revoke(ctx context.Context, id, ownerID uuid.UUID) (*APIKey, error)
}
