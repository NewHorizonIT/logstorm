package apikey_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newMiddlewareRouter(repo apikey.APIKeyRepository) *gin.Engine {
	mw := apikey.NewMiddleware(repo)
	r := gin.New()
	r.GET("/test", mw.ValidateAPIKey, func(c *gin.Context) {
		projectID, ok := apikey.ProjectIDFromContext(c)
		if !ok {
			c.Status(http.StatusInternalServerError)
			return
		}
		c.JSON(http.StatusOK, gin.H{"project_id": projectID.String()})
	})
	return r
}

func TestMiddleware_ValidKey(t *testing.T) {
	projectID := uuid.New()
	key := &apikey.APIKey{
		ID:        uuid.New(),
		ProjectID: projectID,
		KeyHash:   "anyhash",
		KeyPrefix: "lsk_a1b2c3d4",
	}

	repo := &fakeRepo{
		getByHashFn: func(_ context.Context, _ string) (*apikey.APIKey, error) {
			return key, nil
		},
	}

	r := newMiddlewareRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "lsk_anyvalidkey")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, projectID.String(), body["project_id"])
}

func TestMiddleware_MissingHeader(t *testing.T) {
	r := newMiddlewareRouter(&fakeRepo{})
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "MISSING_API_KEY", body["code"])
}

func TestMiddleware_InvalidKey(t *testing.T) {
	repo := &fakeRepo{
		getByHashFn: func(_ context.Context, _ string) (*apikey.APIKey, error) {
			return nil, apikey.ErrKeyNotFound
		},
	}

	r := newMiddlewareRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "lsk_invalidkey")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)

	var body map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &body))
	assert.Equal(t, "INVALID_API_KEY", body["code"])
}

func TestMiddleware_HashIsComputedFromRawKey(t *testing.T) {
	var capturedHash string
	repo := &fakeRepo{
		getByHashFn: func(_ context.Context, keyHash string) (*apikey.APIKey, error) {
			capturedHash = keyHash
			return nil, apikey.ErrKeyNotFound
		},
	}

	r := newMiddlewareRouter(repo)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodGet, "/test", nil)
	req.Header.Set("X-API-Key", "lsk_somerawkey")
	r.ServeHTTP(w, req)

	assert.NotEmpty(t, capturedHash)
	assert.NotEqual(t, "lsk_somerawkey", capturedHash, "hash must differ from raw key")
}
