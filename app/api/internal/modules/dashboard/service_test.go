package dashboard_test

import (
	"context"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/dashboard"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newSvc(repo dashboard.DashboardRepository, projRepo project.ProjectRepository) *dashboard.DashboardService {
	return dashboard.NewDashboardService(repo, projRepo)
}

func stubProjectRepo(ownerID, projectID uuid.UUID) *fakeProjectRepo {
	return &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*project.Project, error) {
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
}

func TestDashboardService_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC()
	from := now.Add(-24 * time.Hour)
	want := &dashboard.LogStats{
		TotalLogs:      10,
		ErrorCount:     2,
		Volume:         []dashboard.VolumePoint{},
		LevelBreakdown: []dashboard.LevelBreakdown{},
		RecentErrors:   []dashboard.RecentError{},
	}

	projRepo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, id, oID uuid.UUID) (*project.Project, error) {
			assert.Equal(t, projectID, id)
			assert.Equal(t, ownerID, oID)
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
	repo := &fakeDashboardRepo{
		getStatsFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) (*dashboard.LogStats, error) {
			return want, nil
		},
	}

	got, err := newSvc(repo, projRepo).GetStats(context.Background(), ownerID, projectID, from, now)

	require.NoError(t, err)
	assert.Equal(t, int64(10), got.TotalLogs)
	assert.Equal(t, int64(2), got.ErrorCount)
}

func TestDashboardService_ForbiddenProject(t *testing.T) {
	t.Parallel()

	projRepo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}

	_, err := newSvc(&fakeDashboardRepo{}, projRepo).GetStats(
		context.Background(), uuid.New(), uuid.New(), time.Time{}, time.Time{},
	)

	assert.ErrorIs(t, err, dashboard.ErrForbidden)
}

func TestDashboardService_DefaultTimeRange(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()

	var capturedFrom, capturedTo time.Time
	repo := &fakeDashboardRepo{
		getStatsFn: func(_ context.Context, _ uuid.UUID, from, to time.Time) (*dashboard.LogStats, error) {
			capturedFrom = from
			capturedTo = to
			return &dashboard.LogStats{}, nil
		},
	}

	_, err := newSvc(repo, stubProjectRepo(ownerID, projectID)).GetStats(
		context.Background(), ownerID, projectID, time.Time{}, time.Time{},
	)

	require.NoError(t, err)
	assert.WithinDuration(t, time.Now().UTC().Add(-24*time.Hour), capturedFrom, 5*time.Second, "from should be ~24h before now")
	assert.WithinDuration(t, time.Now().UTC(), capturedTo, 5*time.Second, "to should be ~now")
}

func TestDashboardService_RepoError(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	repoErr := assert.AnError

	repo := &fakeDashboardRepo{
		getStatsFn: func(_ context.Context, _ uuid.UUID, _, _ time.Time) (*dashboard.LogStats, error) {
			return nil, repoErr
		},
	}

	_, err := newSvc(repo, stubProjectRepo(ownerID, projectID)).GetStats(
		context.Background(), ownerID, projectID, time.Now().Add(-time.Hour), time.Now(),
	)

	assert.ErrorIs(t, err, repoErr)
}
