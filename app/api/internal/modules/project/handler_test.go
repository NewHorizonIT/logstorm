package project_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newHandlerRouterWithRepo exposes the fake repo for test control.
func newHandlerRouterWithRepo(t *testing.T) (*gin.Engine, uuid.UUID, *fakeProjectRepo) {
	t.Helper()

	ownerID := uuid.New()
	repo := &fakeProjectRepo{}
	svc := project.NewProjectService(repo)
	h := project.NewProjectHandler(svc)

	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.UserIDKey, ownerID)
		c.Next()
	})
	project.RegisterRoutes(r.Group("/api/v1"), h)
	return r, ownerID, repo
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func do(r *gin.Engine, method, path string, body *bytes.Buffer) *httptest.ResponseRecorder {
	var req *http.Request
	if body != nil {
		req = httptest.NewRequest(method, path, body)
		req.Header.Set("Content-Type", "application/json")
	} else {
		req = httptest.NewRequest(method, path, nil)
	}
	w := httptest.NewRecorder()
	r.ServeHTTP(w, req)
	return w
}

func parseData(t *testing.T, w *httptest.ResponseRecorder) map[string]any {
	t.Helper()
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok, "response should have 'data' object")
	return data
}

func fixedProj(ownerID uuid.UUID) *project.Project {
	return &project.Project{
		ID:          uuid.New(),
		OwnerID:     ownerID,
		Name:        "My App",
		Slug:        "my-app",
		Environment: "production",
	}
}

// --- Create ---

func TestHandler_Create_Success(t *testing.T) {
	t.Parallel()

	r, ownerID, repo := newHandlerRouterWithRepo(t)
	proj := fixedProj(ownerID)
	repo.createFn = func(_ context.Context, p project.CreateProjectParams) (*project.Project, error) {
		return proj, nil
	}

	w := do(r, http.MethodPost, "/api/v1/projects", jsonBody(t, map[string]string{
		"name": "My App",
	}))

	require.Equal(t, http.StatusCreated, w.Code)
	data := parseData(t, w)
	assert.Equal(t, proj.ID.String(), data["id"])
	assert.Equal(t, "my-app", data["slug"])
}

func TestHandler_Create_MissingName(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodPost, "/api/v1/projects", jsonBody(t, map[string]string{
		"slug": "some-slug",
	}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

func TestHandler_Create_NameTooLong(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodPost, "/api/v1/projects", jsonBody(t, map[string]string{
		"name": strings.Repeat("a", 101),
	}))

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestHandler_Create_InvalidSlug(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodPost, "/api/v1/projects", jsonBody(t, map[string]string{
		"name": "Valid Name",
		"slug": "INVALID SLUG",
	}))

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
	assert.Contains(t, w.Body.String(), "VALIDATION_ERROR")
}

func TestHandler_Create_SlugConflict(t *testing.T) {
	t.Parallel()

	r, _, repo := newHandlerRouterWithRepo(t)
	repo.createFn = func(_ context.Context, _ project.CreateProjectParams) (*project.Project, error) {
		return nil, project.ErrSlugAlreadyExists
	}

	w := do(r, http.MethodPost, "/api/v1/projects", jsonBody(t, map[string]string{
		"name": "My App",
		"slug": "existing-slug",
	}))

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "SLUG_ALREADY_EXISTS")
}

// --- List ---

func TestHandler_List_Success(t *testing.T) {
	t.Parallel()

	r, ownerID, repo := newHandlerRouterWithRepo(t)
	repo.listByOwnerFn = func(_ context.Context, _ uuid.UUID) ([]*project.Project, error) {
		return []*project.Project{fixedProj(ownerID), fixedProj(ownerID)}, nil
	}

	w := do(r, http.MethodGet, "/api/v1/projects", nil)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	arr, ok := resp["data"].([]any)
	require.True(t, ok, "data should be an array")
	assert.Len(t, arr, 2)
}

func TestHandler_List_Empty(t *testing.T) {
	t.Parallel()

	r, _, repo := newHandlerRouterWithRepo(t)
	repo.listByOwnerFn = func(_ context.Context, _ uuid.UUID) ([]*project.Project, error) {
		return []*project.Project{}, nil
	}

	w := do(r, http.MethodGet, "/api/v1/projects", nil)

	require.Equal(t, http.StatusOK, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	arr, ok := resp["data"].([]any)
	require.True(t, ok, "data should be an array")
	assert.Len(t, arr, 0)
}

// --- Get ---

func TestHandler_Get_Success(t *testing.T) {
	t.Parallel()

	r, ownerID, repo := newHandlerRouterWithRepo(t)
	proj := fixedProj(ownerID)
	repo.getByIDFn = func(_ context.Context, id, oID uuid.UUID) (*project.Project, error) {
		assert.Equal(t, proj.ID, id)
		assert.Equal(t, ownerID, oID)
		return proj, nil
	}

	w := do(r, http.MethodGet, "/api/v1/projects/"+proj.ID.String(), nil)

	require.Equal(t, http.StatusOK, w.Code)
	data := parseData(t, w)
	assert.Equal(t, proj.ID.String(), data["id"])
}

func TestHandler_Get_NotFound(t *testing.T) {
	t.Parallel()

	r, _, repo := newHandlerRouterWithRepo(t)
	repo.getByIDFn = func(_ context.Context, _, _ uuid.UUID) (*project.Project, error) {
		return nil, project.ErrProjectNotFound
	}

	w := do(r, http.MethodGet, "/api/v1/projects/"+uuid.New().String(), nil)

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "PROJECT_NOT_FOUND")
}

func TestHandler_Get_InvalidUUID(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodGet, "/api/v1/projects/not-a-uuid", nil)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_PROJECT_ID")
}

// --- Update ---

func TestHandler_Update_Success(t *testing.T) {
	t.Parallel()

	r, ownerID, repo := newHandlerRouterWithRepo(t)
	proj := fixedProj(ownerID)
	repo.updateFn = func(_ context.Context, p project.UpdateProjectParams) (*project.Project, error) {
		assert.Equal(t, proj.ID, p.ID)
		assert.Equal(t, ownerID, p.OwnerID)
		assert.Equal(t, "New Name", p.Name)
		updated := *proj
		updated.Name = p.Name
		return &updated, nil
	}

	w := do(r, http.MethodPut, "/api/v1/projects/"+proj.ID.String(), jsonBody(t, map[string]string{
		"name": "New Name",
	}))

	require.Equal(t, http.StatusOK, w.Code)
	data := parseData(t, w)
	assert.Equal(t, "New Name", data["name"])
}

func TestHandler_Update_EmptyName(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodPut, "/api/v1/projects/"+uuid.New().String(), jsonBody(t, map[string]string{
		"name": "",
	}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

func TestHandler_Update_NotFound(t *testing.T) {
	t.Parallel()

	r, _, repo := newHandlerRouterWithRepo(t)
	repo.updateFn = func(_ context.Context, _ project.UpdateProjectParams) (*project.Project, error) {
		return nil, project.ErrProjectNotFound
	}

	w := do(r, http.MethodPut, "/api/v1/projects/"+uuid.New().String(), jsonBody(t, map[string]string{
		"name": "Valid Name",
	}))

	assert.Equal(t, http.StatusNotFound, w.Code)
	assert.Contains(t, w.Body.String(), "PROJECT_NOT_FOUND")
}

func TestHandler_Update_InvalidUUID(t *testing.T) {
	t.Parallel()

	r, _, _ := newHandlerRouterWithRepo(t)

	w := do(r, http.MethodPut, "/api/v1/projects/bad-id", jsonBody(t, map[string]string{
		"name": "Valid Name",
	}))

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_PROJECT_ID")
}
