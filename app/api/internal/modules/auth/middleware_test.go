package auth_test

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// testMiddleware returns an auth.Middleware wired with the shared test TokenService.
func testMiddleware() auth.Middleware {
	return auth.NewMiddleware(testTokenService())
}

func TestMiddleware_Authenticate_NoHeader(t *testing.T) {
	t.Parallel()

	mw := testMiddleware()
	r := gin.New()
	r.GET("/protected", mw.Authenticate, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/protected", nil))

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_TOKEN")
}

func TestMiddleware_Authenticate_MalformedHeader(t *testing.T) {
	t.Parallel()

	mw := testMiddleware()
	r := gin.New()
	r.GET("/protected", mw.Authenticate, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	// "bearer" lowercase, "Token <token>", bare token — all should be rejected
	cases := []string{"bearer sometoken", "Token sometoken", "justtoken"}
	for _, header := range cases {
		w := httptest.NewRecorder()
		req := httptest.NewRequest(http.MethodGet, "/protected", nil)
		req.Header.Set("Authorization", header)
		r.ServeHTTP(w, req)
		assert.Equal(t, http.StatusUnauthorized, w.Code, "header: %q", header)
	}
}

func TestMiddleware_Authenticate_InvalidToken(t *testing.T) {
	t.Parallel()

	mw := testMiddleware()
	r := gin.New()
	r.GET("/protected", mw.Authenticate, func(c *gin.Context) {
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer invalid.token.here")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TOKEN")
}

func TestMiddleware_Authenticate_Valid(t *testing.T) {
	t.Parallel()

	tokenSvc := testTokenService()
	mw := auth.NewMiddleware(tokenSvc)

	userID := uuid.New()
	token, err := tokenSvc.GenerateAccessToken(userID)
	require.NoError(t, err)

	var gotID uuid.UUID
	var gotOK bool

	r := gin.New()
	r.GET("/protected", mw.Authenticate, func(c *gin.Context) {
		gotID, gotOK = auth.UserIDFromContext(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/protected", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusOK, w.Code)
	assert.True(t, gotOK)
	assert.Equal(t, userID, gotID)
}

func TestUserIDFromContext_Present(t *testing.T) {
	t.Parallel()

	tokenSvc := testTokenService()
	mw := auth.NewMiddleware(tokenSvc)

	userID := uuid.New()
	token, err := tokenSvc.GenerateAccessToken(userID)
	require.NoError(t, err)

	var got uuid.UUID
	var ok bool

	r := gin.New()
	r.GET("/", mw.Authenticate, func(c *gin.Context) {
		got, ok = auth.UserIDFromContext(c)
		c.Status(http.StatusOK)
	})

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodGet, "/", nil)
	req.Header.Set("Authorization", "Bearer "+token)
	r.ServeHTTP(w, req)

	assert.True(t, ok)
	assert.Equal(t, userID, got)
}

func TestUserIDFromContext_Missing(t *testing.T) {
	t.Parallel()

	var got uuid.UUID
	var ok bool

	r := gin.New()
	r.GET("/", func(c *gin.Context) {
		got, ok = auth.UserIDFromContext(c)
		c.Status(http.StatusOK)
	})

	httptest.NewRecorder()
	w := httptest.NewRecorder()
	r.ServeHTTP(w, httptest.NewRequest(http.MethodGet, "/", nil))

	assert.False(t, ok)
	assert.Equal(t, uuid.Nil, got)
}
