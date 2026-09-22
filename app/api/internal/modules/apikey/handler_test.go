package apikey_test

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

func newHandlerRouter(t *testing.T, repo apikey.APIKeyRepository, pv apikey.ProjectVerifier) *gin.Engine {
	t.Helper()
	svc := apikey.NewAPIKeyService(repo, pv)
	h := apikey.NewAPIKeyHandler(svc)
	r := gin.New()
	g := r.Group("/api/v1")
	g.Use(func(c *gin.Context) {
		c.Set(auth.UserIDKey, uuid.MustParse("00000000-0000-0000-0000-000000000001"))
		c.Next()
	})
	apikey.RegisterRoutes(g, h)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func fixedKeyWithTime(ownerID, projectID uuid.UUID, name string) *apikey.APIKey {
	return &apikey.APIKey{
		ID:        uuid.New(),
		OwnerID:   ownerID,
		ProjectID: projectID,
		Name:      name,
		KeyPrefix: "lsk_a1b2c3d4",
		KeyHash:   "fakehash",
		CreatedAt: time.Now(),
	}
}

// --- Create ---

func TestHandler_Create_Success(t *testing.T) {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")
	projectID := uuid.New()

	repo := &fakeRepo{
		createFn: func(_ context.Context, p apikey.CreateAPIKeyParams) (*apikey.APIKey, error) {
			return fixedKeyWithTime(ownerID, projectID, p.Name), nil
		},
	}

	r := newHandlerRouter(t, repo, okVerifier())
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]string{"project_id": projectID.String(), "name": "My Key"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/api-keys", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].(map[string]any)
	assert.NotEmpty(t, data["key"], "raw key must be present in create response")
	assert.NotEmpty(t, data["key_prefix"])
	assert.Equal(t, "My Key", data["name"])
}

func TestHandler_Create_MissingName(t *testing.T) {
	r := newHandlerRouter(t, &fakeRepo{}, okVerifier())
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]string{"project_id": uuid.New().String()})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/api-keys", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Create_InvalidProjectID(t *testing.T) {
	r := newHandlerRouter(t, &fakeRepo{}, okVerifier())
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]string{"project_id": "not-a-uuid", "name": "Key"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/api-keys", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestHandler_Create_ProjectNotFound(t *testing.T) {
	r := newHandlerRouter(t, &fakeRepo{}, notFoundVerifier())
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]string{"project_id": uuid.New().String(), "name": "Key"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/api-keys", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

// --- List ---

func TestHandler_List_Success(t *testing.T) {
	ownerID := uuid.MustParse("00000000-0000-0000-0000-000000000001")

	repo := &fakeRepo{
		listByOwnerFn: func(_ context.Context, _ uuid.UUID) ([]*apikey.APIKey, error) {
			return []*apikey.APIKey{
				fixedKeyWithTime(ownerID, uuid.New(), "A"),
				fixedKeyWithTime(ownerID, uuid.New(), "B"),
			}, nil
		},
	}

	r := newHandlerRouter(t, repo, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/api-keys", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data := resp["data"].([]any)
	assert.Len(t, data, 2)

	for _, item := range data {
		m := item.(map[string]any)
		assert.Nil(t, m["key"], "raw key must NOT appear in list response")
		assert.NotEmpty(t, m["key_prefix"])
	}
}

func TestHandler_List_FilterByProject(t *testing.T) {
	projectID := uuid.New()

	repo := &fakeRepo{
		listByProjectFn: func(_ context.Context, pID, _ uuid.UUID) ([]*apikey.APIKey, error) {
			assert.Equal(t, projectID, pID)
			return []*apikey.APIKey{}, nil
		},
	}

	r := newHandlerRouter(t, repo, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/api-keys?project_id="+projectID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
}

func TestHandler_List_InvalidProjectIDParam(t *testing.T) {
	r := newHandlerRouter(t, &fakeRepo{}, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/api/v1/api-keys?project_id=bad-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

// --- Revoke ---

func TestHandler_Revoke_Success(t *testing.T) {
	keyID := uuid.New()

	repo := &fakeRepo{
		revokeFn: func(_ context.Context, id, _ uuid.UUID) (*apikey.APIKey, error) {
			assert.Equal(t, keyID, id)
			return fixedKeyWithTime(uuid.New(), uuid.New(), "key"), nil
		},
	}

	r := newHandlerRouter(t, repo, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+keyID.String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

func TestHandler_Revoke_NotFound(t *testing.T) {
	repo := &fakeRepo{
		revokeFn: func(_ context.Context, _, _ uuid.UUID) (*apikey.APIKey, error) {
			return nil, apikey.ErrKeyNotFound
		},
	}

	r := newHandlerRouter(t, repo, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/api-keys/"+uuid.New().String(), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNotFound, w.Code)
}

func TestHandler_Revoke_InvalidID(t *testing.T) {
	r := newHandlerRouter(t, &fakeRepo{}, okVerifier())
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodDelete, "/api/v1/api-keys/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}
