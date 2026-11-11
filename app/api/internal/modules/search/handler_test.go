package search_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newSearchRouter builds a test router with a fixed ownerID injected into context.
func newSearchRouter(t *testing.T, ownerID uuid.UUID, projRepo *fakeProjectRepo, logRepo *fakeLogRepo) *gin.Engine {
	t.Helper()
	svc := search.NewSearchService(logRepo, projRepo)
	h := search.NewSearchHandler(svc)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.UserIDKey, ownerID)
		c.Next()
	})
	search.RegisterSearchRoutes(r.Group("/api/v1"), h)
	return r
}

// newSearchRouterNoAuth builds a router without auth middleware.
func newSearchRouterNoAuth(t *testing.T, projRepo *fakeProjectRepo, logRepo *fakeLogRepo) *gin.Engine {
	t.Helper()
	svc := search.NewSearchService(logRepo, projRepo)
	h := search.NewSearchHandler(svc)
	r := gin.New()
	search.RegisterSearchRoutes(r.Group("/api/v1"), h)
	return r
}

func searchURL(projectID uuid.UUID, query string) string {
	u := "/api/v1/projects/" + projectID.String() + "/logs"
	if query != "" {
		u += "?" + query
	}
	return u
}

func doGet(r *gin.Engine, path string) *httptest.ResponseRecorder {
	req := httptest.NewRequest(http.MethodGet, path, nil)
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func okProjRepo(ownerID, projectID uuid.UUID) *fakeProjectRepo {
	return &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*project.Project, error) {
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
}

// --- tests ---

func TestSearchHandler_Success(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	logID := uuid.New()

	logRepo := &fakeLogRepo{
		searchFn: func(_ context.Context, q *search.SearchQuery) (*search.SearchResult, error) {
			return &search.SearchResult{
				Logs:  []*search.LogRecord{{ID: logID, Message: "hello world"}},
				Total: 1,
			}, nil
		},
	}
	r := newSearchRouter(t, ownerID, okProjRepo(ownerID, projectID), logRepo)

	w := doGet(r, searchURL(projectID, ""))

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.Equal(t, float64(1), data["total"])
	logs := data["logs"].([]any)
	assert.Len(t, logs, 1)
}

func TestSearchHandler_EmptyResults_ReturnsArray(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()

	r := newSearchRouter(t, ownerID, okProjRepo(ownerID, projectID), &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, ""))

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)

	// logs must be [] not null
	raw, _ := json.Marshal(data["logs"])
	assert.Equal(t, "[]", string(raw), "empty logs must serialize as [] not null")
}

func TestSearchHandler_MissingAuth_Returns401(t *testing.T) {
	t.Parallel()

	projectID := uuid.New()
	r := newSearchRouterNoAuth(t, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, ""))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_TOKEN")
}

func TestSearchHandler_InvalidProjectID_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	r := newSearchRouter(t, ownerID, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, "/api/v1/projects/not-a-uuid/logs")

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_PROJECT_ID")
}

func TestSearchHandler_InvalidFrom_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	r := newSearchRouter(t, ownerID, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "from=not-a-date"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_FROM_TIME")
}

func TestSearchHandler_InvalidTo_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	r := newSearchRouter(t, ownerID, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "to=bad-time"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TO_TIME")
}

func TestSearchHandler_InvalidLimit_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	r := newSearchRouter(t, ownerID, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "limit=abc"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_LIMIT")
}

func TestSearchHandler_InvalidOffset_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	r := newSearchRouter(t, ownerID, &fakeProjectRepo{}, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "offset=xyz"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_OFFSET")
}

func TestSearchHandler_ForbiddenProject_Returns403(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	projRepo := &fakeProjectRepo{
		getByIDFn: func(_ context.Context, _, _ uuid.UUID) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}
	r := newSearchRouter(t, ownerID, projRepo, &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, ""))

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "FORBIDDEN")
}

func TestSearchHandler_ValidationError_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	r := newSearchRouter(t, ownerID, okProjRepo(ownerID, projectID), &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "limit=1001"))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestSearchHandler_FromAfterTo_Returns400(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	now := time.Now().UTC()
	from := now.Format(time.RFC3339)
	to := now.Add(-time.Hour).Format(time.RFC3339)
	r := newSearchRouter(t, ownerID, okProjRepo(ownerID, projectID), &fakeLogRepo{})

	w := doGet(r, searchURL(projectID, "from="+from+"&to="+to))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestSearchHandler_LogRepoError_Returns500(t *testing.T) {
	t.Parallel()

	ownerID := uuid.New()
	projectID := uuid.New()
	logRepo := &fakeLogRepo{
		searchFn: func(_ context.Context, _ *search.SearchQuery) (*search.SearchResult, error) {
			return nil, errors.New("clickhouse down")
		},
	}
	r := newSearchRouter(t, ownerID, okProjRepo(ownerID, projectID), logRepo)

	w := doGet(r, searchURL(projectID, ""))

	assert.Equal(t, http.StatusInternalServerError, w.Code)
	assert.Contains(t, w.Body.String(), "INTERNAL_SERVER_ERROR")
}
