package ingestion_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/logstorm/api/internal/modules/ingestion"
)

// --- fake publisher ---

type fakePublisher struct {
	publishFn func(ctx context.Context, event *ingestion.LogEvent) error
}

func (f *fakePublisher) Publish(ctx context.Context, event *ingestion.LogEvent) error {
	if f.publishFn != nil {
		return f.publishFn(ctx, event)
	}
	return nil
}

var fixedProjectID = uuid.MustParse("00000000-0000-0000-0000-000000000001")

// newHandlerRouter sets up a router for handler unit tests.
// withAPIKey=true injects a projectID into the context, simulating a valid API key.
func newHandlerRouter(t *testing.T, pub ingestion.Publisher, withAPIKey bool) *gin.Engine {
	t.Helper()
	h := ingestion.NewIngestionHandler(pub)
	r := gin.New()
	g := r.Group("/api/v1")
	if withAPIKey {
		g.Use(func(c *gin.Context) {
			c.Set(apikey.ProjectIDKey, fixedProjectID)
			c.Next()
		})
	}
	g.POST("/logs", h.Ingest)
	return r
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

// --- tests ---

func TestIngest_Success(t *testing.T) {
	var capturedEvent *ingestion.LogEvent
	pub := &fakePublisher{
		publishFn: func(_ context.Context, e *ingestion.LogEvent) error {
			capturedEvent = e
			return nil
		},
	}

	r := newHandlerRouter(t, pub, true)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{
		"level":   "info",
		"message": "user signed in",
		"service": "auth",
	})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusAccepted, w.Code)

	require.NotNil(t, capturedEvent)
	assert.Equal(t, fixedProjectID, capturedEvent.ProjectID)
	assert.Equal(t, "info", capturedEvent.Level)
	assert.Equal(t, "sdk", capturedEvent.Source)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.NotEmpty(t, resp["data"].(map[string]any)["id"])
}

func TestIngest_MissingAPIKey_Returns401(t *testing.T) {
	r := newHandlerRouter(t, &fakePublisher{}, false)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{"level": "info", "message": "msg"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
}

func TestIngest_InvalidJSON_Returns400(t *testing.T) {
	r := newHandlerRouter(t, &fakePublisher{}, true)
	w := httptest.NewRecorder()
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", bytes.NewBufferString(`{invalid}`))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
}

func TestIngest_InvalidLevel_Returns422(t *testing.T) {
	r := newHandlerRouter(t, &fakePublisher{}, true)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{"level": "critical", "message": "msg"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestIngest_EmptyMessage_Returns422(t *testing.T) {
	r := newHandlerRouter(t, &fakePublisher{}, true)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{"level": "info", "message": "   "})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnprocessableEntity, w.Code)
}

func TestIngest_PublishError_Returns500(t *testing.T) {
	pub := &fakePublisher{
		publishFn: func(_ context.Context, _ *ingestion.LogEvent) error {
			return errors.New("broker unavailable")
		},
	}

	r := newHandlerRouter(t, pub, true)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{"level": "error", "message": "something failed"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusInternalServerError, w.Code)
}

func TestIngest_ErrorResponse_HasCodeAndMessage(t *testing.T) {
	r := newHandlerRouter(t, &fakePublisher{}, true)
	w := httptest.NewRecorder()
	body := jsonBody(t, map[string]any{"level": "bad", "message": "msg"})
	req, _ := http.NewRequest(http.MethodPost, "/api/v1/logs", body)
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	assert.Equal(t, "VALIDATION_ERROR", resp["code"])
	assert.NotEmpty(t, resp["message"])
}
