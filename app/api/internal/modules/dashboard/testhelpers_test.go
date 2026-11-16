package dashboard_test

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/dashboard"
	"github.com/logstorm/api/internal/modules/project"
)

type fakeDashboardRepo struct {
	getStatsFn func(ctx context.Context, projectID uuid.UUID, from, to time.Time) (*dashboard.LogStats, error)
}

func (f *fakeDashboardRepo) GetStats(ctx context.Context, projectID uuid.UUID, from, to time.Time) (*dashboard.LogStats, error) {
	if f.getStatsFn != nil {
		return f.getStatsFn(ctx, projectID, from, to)
	}
	return &dashboard.LogStats{
		Volume:         []dashboard.VolumePoint{},
		LevelBreakdown: []dashboard.LevelBreakdown{},
		RecentErrors:   []dashboard.RecentError{},
	}, nil
}

type fakeProjectRepo struct {
	getByIDFn func(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error)
}

func (f *fakeProjectRepo) Create(_ context.Context, _ project.CreateProjectParams) (*project.Project, error) {
	panic("not implemented")
}
func (f *fakeProjectRepo) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error) {
	return f.getByIDFn(ctx, id, ownerID)
}
func (f *fakeProjectRepo) GetByOwnerAndSlug(_ context.Context, _ uuid.UUID, _ string) (*project.Project, error) {
	panic("not implemented")
}
func (f *fakeProjectRepo) ListByOwner(_ context.Context, _ uuid.UUID) ([]*project.Project, error) {
	panic("not implemented")
}
func (f *fakeProjectRepo) Update(_ context.Context, _ project.UpdateProjectParams) (*project.Project, error) {
	panic("not implemented")
}
