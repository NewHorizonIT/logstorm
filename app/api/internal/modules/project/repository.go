package project

import (
	"context"

	"github.com/google/uuid"
)

type CreateProjectParams struct {
	OwnerID     uuid.UUID
	Name        string
	Slug        string
	Description string
	Environment string
}

type UpdateProjectParams struct {
	ID          uuid.UUID
	OwnerID     uuid.UUID
	Name        string
	Description string
}

type ProjectRepository interface {
	Create(ctx context.Context, params CreateProjectParams) (*Project, error)
	GetByID(ctx context.Context, id, ownerID uuid.UUID) (*Project, error)
	GetByOwnerAndSlug(ctx context.Context, ownerID uuid.UUID, slug string) (*Project, error)
	ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*Project, error)
	Update(ctx context.Context, params UpdateProjectParams) (*Project, error)
}
