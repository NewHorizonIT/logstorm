package search_test

import (
	"context"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
)

type fakeLogRepo struct {
	searchFn func(ctx context.Context, q *search.SearchQuery) (*search.SearchResult, error)
}

func (f *fakeLogRepo) Search(ctx context.Context, q *search.SearchQuery) (*search.SearchResult, error) {
	if f.searchFn != nil {
		return f.searchFn(ctx, q)
	}
	return &search.SearchResult{Logs: make([]*search.LogRecord, 0)}, nil
}

type fakeProjectRepo struct {
	getByIDFn           func(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error)
	createFn            func(ctx context.Context, params project.CreateProjectParams) (*project.Project, error)
	getByOwnerAndSlugFn func(ctx context.Context, ownerID uuid.UUID, slug string) (*project.Project, error)
	listByOwnerFn       func(ctx context.Context, ownerID uuid.UUID) ([]*project.Project, error)
	updateFn            func(ctx context.Context, params project.UpdateProjectParams) (*project.Project, error)
}

func (f *fakeProjectRepo) Create(ctx context.Context, params project.CreateProjectParams) (*project.Project, error) {
	return f.createFn(ctx, params)
}
func (f *fakeProjectRepo) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error) {
	return f.getByIDFn(ctx, id, ownerID)
}
func (f *fakeProjectRepo) GetByOwnerAndSlug(ctx context.Context, ownerID uuid.UUID, slug string) (*project.Project, error) {
	return f.getByOwnerAndSlugFn(ctx, ownerID, slug)
}
func (f *fakeProjectRepo) ListByOwner(ctx context.Context, ownerID uuid.UUID) ([]*project.Project, error) {
	return f.listByOwnerFn(ctx, ownerID)
}
func (f *fakeProjectRepo) Update(ctx context.Context, params project.UpdateProjectParams) (*project.Project, error) {
	return f.updateFn(ctx, params)
}
