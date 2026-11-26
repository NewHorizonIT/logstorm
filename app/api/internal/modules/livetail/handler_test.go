package livetail_test

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/livetail"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
	"github.com/stretchr/testify/assert"
)

func init() {
	gin.SetMode(gin.TestMode)
}

type fakeSubscriber struct {
	ch          chan *search.LogRecord
	unsubCalled bool
}

func newFakeSubscriber() *fakeSubscriber {
	return &fakeSubscriber{ch: make(chan *search.LogRecord, livetail.SubscriberBufferSize)}
}

func (f *fakeSubscriber) Subscribe(_ uuid.UUID, _ livetail.Filter) (<-chan *search.LogRecord, func()) {
	return f.ch, func() { f.unsubCalled = true }
}

type fakeHandlerProjectRepo struct {
	getByIDFn func(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error)
}

func (f *fakeHandlerProjectRepo) GetByID(ctx context.Context, id, ownerID uuid.UUID) (*project.Project, error) {
	return f.getByIDFn(ctx, id, ownerID)
}
func (f *fakeHandlerProjectRepo) Create(_ context.Context, _ project.CreateProjectParams) (*project.Project, error) {
	return nil, nil
}
func (f *fakeHandlerProjectRepo) GetByOwnerAndSlug(_ context.Context, _ uuid.UUID, _ string) (*project.Project, error) {
	return nil, nil
}
func (f *fakeHandlerProjectRepo) ListByOwner(_ context.Context, _ uuid.UUID) ([]*project.Project, error) {
	return nil, nil
}
func (f *fakeHandlerProjectRepo) Update(_ context.Context, _ project.UpdateProjectParams) (*project.Project, error) {
	return nil, nil
}

func okProjectRepo(ownerID, projectID uuid.UUID) *fakeHandlerProjectRepo {
	return &fakeHandlerProjectRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*project.Project, error) {
			return &project.Project{ID: projectID, OwnerID: ownerID}, nil
		},
	}
}

func notFoundProjectRepo() *fakeHandlerProjectRepo {
	return &fakeHandlerProjectRepo{
		getByIDFn: func(_ context.Context, _ uuid.UUID, _ uuid.UUID) (*project.Project, error) {
			return nil, project.ErrProjectNotFound
		},
	}
}

// --- router helpers ---

// buildHandlerRouter tạo test router với ownerID được inject vào context.
// Dùng NewHandlerForTest để nhận fake Subscriber thay vì *Hub thật.
func buildHandlerRouter(ownerID uuid.UUID, hub livetail.Subscriber, projRepo project.ProjectRepository) *gin.Engine {
	h := livetail.NewHandlerForTest(hub, projRepo)
	r := gin.New()
	r.Use(func(c *gin.Context) {
		c.Set(auth.UserIDKey, ownerID)
		c.Next()
	})
	r.GET("/api/v1/livetail/:project_id", h.Stream)
	return r
}

// buildHandlerRouterNoAuth tạo test router không inject auth — dùng để test 401.
func buildHandlerRouterNoAuth(hub livetail.Subscriber, projRepo project.ProjectRepository) *gin.Engine {
	h := livetail.NewHandlerForTest(hub, projRepo)
	r := gin.New()
	r.GET("/api/v1/livetail/:project_id", h.Stream)
	return r
}

func tailURL(projectID uuid.UUID) string {
	return "/api/v1/livetail/" + projectID.String()
}

// --- tests ---

// TestHandler_MissingAuth_Returns401 kiểm tra rằng request không có JWT
// nhận 401 với code MISSING_TOKEN.
func TestHandler_MissingAuth_Returns401(t *testing.T) {
	hub := newFakeSubscriber()
	r := buildHandlerRouterNoAuth(hub, okProjectRepo(uuid.New(), uuid.New()))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, tailURL(uuid.New()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_TOKEN")
}

// TestHandler_InvalidProjectID_Returns400 kiểm tra rằng project_id không phải UUID
// nhận 400 với code INVALID_PROJECT_ID.
func TestHandler_InvalidProjectID_Returns400(t *testing.T) {
	hub := newFakeSubscriber()
	ownerID := uuid.New()
	r := buildHandlerRouter(ownerID, hub, okProjectRepo(ownerID, uuid.New()))

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/api/v1/livetail/not-a-uuid", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_PROJECT_ID")
}

// TestHandler_ProjectNotFound_Returns403 kiểm tra rằng project không thuộc user
// nhận 403 với code FORBIDDEN.
func TestHandler_ProjectNotFound_Returns403(t *testing.T) {
	hub := newFakeSubscriber()
	ownerID := uuid.New()
	r := buildHandlerRouter(ownerID, hub, notFoundProjectRepo())

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, tailURL(uuid.New()), nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusForbidden, w.Code)
	assert.Contains(t, w.Body.String(), "FORBIDDEN")
}

// TestHandler_EventReceived_SSEBodyContainsLog kiểm tra rằng khi Hub gửi event,
// response body chứa "event: log" và data của record.
func TestHandler_EventReceived_SSEBodyContainsLog(t *testing.T) {
	hub := newFakeSubscriber()
	ownerID := uuid.New()
	projectID := uuid.New()

	// Pre-populate channel với 1 event trước khi handler start
	hub.ch <- &search.LogRecord{
		ID:      uuid.New(),
		Message: "hello from live tail",
		Level:   "info",
	}

	ctx, cancel := context.WithCancel(context.Background())

	r := buildHandlerRouter(ownerID, hub, okProjectRepo(ownerID, projectID))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, tailURL(projectID), nil)
	req = req.WithContext(ctx)

	done := make(chan struct{})
	go func() {
		r.ServeHTTP(w, req)
		close(done)
	}()

	// Chờ handler xử lý event rồi simulate client disconnect
	time.Sleep(100 * time.Millisecond)
	cancel()
	<-done

	assert.Contains(t, w.Body.String(), "event: log")
	assert.Contains(t, w.Body.String(), "hello from live tail")
}

func TestHandler_ClientDisconnect_UnsubCalled(t *testing.T) {
	hub := newFakeSubscriber()
	ownerID := uuid.New()
	projectID := uuid.New()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel ngay lập tức để simulate client disconnect

	r := buildHandlerRouter(ownerID, hub, okProjectRepo(ownerID, projectID))
	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, tailURL(projectID), nil)
	req = req.WithContext(ctx)

	r.ServeHTTP(w, req)

	assert.True(t, hub.unsubCalled, "unsub phải được gọi khi client disconnect")
}
