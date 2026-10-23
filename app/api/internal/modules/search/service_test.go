package search_test

import (
	"context"
	"errors"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- fake repos ---

type fakeLogRepo struct {
	searchFn func(ctx context.Context, q *search.SearchQuery) (*search.SearchResult, error)
}

func (f *fakeLogRepo) Search(ctx context.Context, q *search.SearchQuery) (*search.SearchResult, error) {
	return f.searchFn(ctx, q)
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

// --- helpers ---

func newSearchSvc(logRepo search.LogRepository, projRepo project.ProjectRepository) *search.SearchService {
	return search.NewSearchService(logRepo, projRepo)
}

func validQuery(projectID uuid.UUID) *search.SearchQuery {
	now := time.Now().UTC()
	return &search.SearchQuery{
		ProjectID: projectID,
		From:      now.Add(-time.Hour),
		To:        now,
		Limit:     50,
	}
}

func stubProjectRepo(ownerID, projectID uuid.UUID) *fakeProjectRepo {
	return &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*project.Project, error) {
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
}

// --- tests ---

func TestSearchService_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	want := &search.SearchResult{
		Logs:  []*search.LogRecord{{ID: uuid.New(), Message: "hello"}},
		Total: 1,
	}
	projRepo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, id, oID uuid.UUID) (*project.Project, error) {
			assert.Equal(t, projectID, id)
			assert.Equal(t, ownerID, oID)
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
	logRepo := &fakeLogRepo{
		searchFn: func(_ context.Context, _ *search.SearchQuery) (*search.SearchResult, error) {
			return want, nil
		},
	}

	got, err := newSearchSvc(logRepo, projRepo).SearchLogs(context.Background(), ownerID, validQuery(projectID))

	require.NoError(t, err)
	assert.Equal(t, int64(1), got.Total)
	assert.Len(t, got.Logs, 1)
}

func TestSearchService_ForbiddenProject(t *testing.T) {
	t.Parallel()

	projRepo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}

	_, err := newSearchSvc(&fakeLogRepo{}, projRepo).SearchLogs(context.Background(), uuid.New(), validQuery(uuid.New()))

	assert.ErrorIs(t, err, search.ErrForbidden)
}

func TestSearchService_ValidatesQuery(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()

	q := &search.SearchQuery{
		ProjectID: projectID,
		Limit:     1001, // exceeds maximum
	}

	_, err := newSearchSvc(&fakeLogRepo{}, stubProjectRepo(ownerID, projectID)).
		SearchLogs(context.Background(), ownerID, q)

	assert.ErrorIs(t, err, search.ErrValidation)
}

func TestSearchService_BadInput(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC()

	q := &search.SearchQuery{
		ProjectID: projectID,
		From:      now,
		To:        now.Add(-time.Hour), // from is after to
		Limit:     50,
	}

	_, err := newSearchSvc(&fakeLogRepo{}, stubProjectRepo(ownerID, projectID)).
		SearchLogs(context.Background(), ownerID, q)

	assert.ErrorIs(t, err, search.ErrValidation)
}

func TestSearchService_ErrorFromLogRepo(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	repoErr := errors.New("clickhouse unavailable")

	logRepo := &fakeLogRepo{
		searchFn: func(_ context.Context, _ *search.SearchQuery) (*search.SearchResult, error) {
			return nil, repoErr
		},
	}

	_, err := newSearchSvc(logRepo, stubProjectRepo(ownerID, projectID)).
		SearchLogs(context.Background(), ownerID, validQuery(projectID))

	assert.ErrorIs(t, err, repoErr)
}
