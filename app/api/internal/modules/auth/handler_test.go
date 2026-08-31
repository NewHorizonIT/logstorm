package auth_test

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
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/user"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func init() {
	gin.SetMode(gin.TestMode)
}

// newHandlerRouter wires up a real AuthService (with fakes injected) and returns
// the router, the user-provider fake, and the auth-repo fake for test control.
func newHandlerRouter(t *testing.T) (*gin.Engine, *fakeUserProvider, *fakeAuthRepo) {
	t.Helper()

	userProv := &fakeUserProvider{}
	authRepo := &fakeAuthRepo{}
	tokenSvc := auth.NewTokenService(config.AuthConfig{
		JWTSecret:       "test-secret-key-minimum-32-chars!!",
		AccessTokenTTL:  15 * time.Minute,
		RefreshTokenTTL: 7 * 24 * time.Hour,
	})
	authSvc := auth.NewAuthService(userProv, authRepo, tokenSvc)
	cfg := &config.Config{App: config.AppConfig{Environment: "development"}}
	handler := auth.NewAuthHandler(authSvc, tokenSvc, cfg)

	r := gin.New()
	auth.RegisterRoutes(r.Group("/api/v1"), handler)
	return r, userProv, authRepo
}

func jsonBody(t *testing.T, v any) *bytes.Buffer {
	t.Helper()
	b, err := json.Marshal(v)
	require.NoError(t, err)
	return bytes.NewBuffer(b)
}

func findCookie(w *httptest.ResponseRecorder, name string) *http.Cookie {
	for _, c := range w.Result().Cookies() {
		if c.Name == name {
			return c
		}
	}
	return nil
}

// --- Register ---

func TestHandler_Register_Success(t *testing.T) {
	t.Parallel()
	r, userProv, _ := newHandlerRouter(t)

	want := &user.User{ID: uuid.New(), Email: "new@example.com", FullName: "New User", Status: "active"}
	userProv.createUserFn = func(_ context.Context, _ user.CreateUserInput) (*user.User, error) {
		return want, nil
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":     "new@example.com",
		"full_name": "New User",
		"password":  "securepass",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusCreated, w.Code)
	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok, "response should have 'data' key")
	assert.Equal(t, want.ID.String(), data["id"])
	assert.Equal(t, "new@example.com", data["email"])
}

func TestHandler_Register_MissingFields(t *testing.T) {
	t.Parallel()
	r, _, _ := newHandlerRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email": "no-password@example.com",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

func TestHandler_Register_PasswordTooShort(t *testing.T) {
	t.Parallel()
	r, _, _ := newHandlerRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":     "user@example.com",
		"full_name": "User",
		"password":  "short",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

func TestHandler_Register_DuplicateEmail(t *testing.T) {
	t.Parallel()
	r, userProv, _ := newHandlerRouter(t)

	userProv.createUserFn = func(_ context.Context, _ user.CreateUserInput) (*user.User, error) {
		return nil, user.ErrUserEmailAlreadyExists
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/register", jsonBody(t, map[string]string{
		"email":     "dup@example.com",
		"full_name": "Dup User",
		"password":  "password123",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusConflict, w.Code)
	assert.Contains(t, w.Body.String(), "EMAIL_ALREADY_EXISTS")
}

// --- Login ---

func TestHandler_Login_Success(t *testing.T) {
	t.Parallel()
	r, userProv, authRepo := newHandlerRouter(t)

	u := &user.User{ID: uuid.New(), PasswordHash: hashedPassword(t, "password123")}
	userProv.getByEmailFn = func(_ context.Context, _ string) (*user.User, error) { return u, nil }
	authRepo.createFn = func(_ context.Context, _ auth.CreateRefreshTokenParams) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{ID: uuid.New(), UserID: u.ID}, nil
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
		"email":    "user@example.com",
		"password": "password123",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok, "response should have 'data' key")
	assert.NotEmpty(t, data["access_token"])
	assert.Equal(t, "Bearer", data["token_type"])
	assert.Greater(t, data["expires_in"], float64(0))

	cookie := findCookie(w, "refresh_token")
	require.NotNil(t, cookie, "refresh_token cookie should be set")
	assert.True(t, cookie.HttpOnly)
	assert.NotEmpty(t, cookie.Value)
}

func TestHandler_Login_InvalidCredentials(t *testing.T) {
	t.Parallel()
	r, userProv, _ := newHandlerRouter(t)

	userProv.getByEmailFn = func(_ context.Context, _ string) (*user.User, error) {
		return nil, user.ErrUserNotFound
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
		"email":    "ghost@example.com",
		"password": "anything",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_CREDENTIALS")
}

func TestHandler_Login_MissingFields(t *testing.T) {
	t.Parallel()
	r, _, _ := newHandlerRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/login", jsonBody(t, map[string]string{
		"email": "user@example.com",
	}))
	req.Header.Set("Content-Type", "application/json")
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_REQUEST")
}

// --- Logout ---

func TestHandler_Logout_WithCookie(t *testing.T) {
	t.Parallel()
	r, _, authRepo := newHandlerRouter(t)

	authRepo.revokeFn = func(_ context.Context, _ string) error { return nil }

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "some-raw-token"})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)

	cookie := findCookie(w, "refresh_token")
	require.NotNil(t, cookie, "cleared refresh_token cookie should be present in response")
	assert.Empty(t, cookie.Value)
	assert.Equal(t, -1, cookie.MaxAge) // Go parses Max-Age=0 header back as -1 (delete semantics)
}

func TestHandler_Logout_WithoutCookie(t *testing.T) {
	t.Parallel()
	r, _, _ := newHandlerRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/logout", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusNoContent, w.Code)
}

// --- Refresh ---

func TestHandler_Refresh_Success(t *testing.T) {
	t.Parallel()
	r, _, authRepo := newHandlerRouter(t)
	userID := uuid.New()

	authRepo.getByHashFn = func(_ context.Context, _ string) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{
			ID:        uuid.New(),
			UserID:    userID,
			ExpiresAt: time.Now().Add(7 * 24 * time.Hour),
		}, nil
	}
	authRepo.rotateFn = func(_ context.Context, _ string, _ auth.CreateRefreshTokenParams) (*auth.RefreshToken, error) {
		return &auth.RefreshToken{ID: uuid.New(), UserID: userID}, nil
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "old-raw-token"})
	r.ServeHTTP(w, req)

	require.Equal(t, http.StatusOK, w.Code)

	var resp map[string]any
	require.NoError(t, json.Unmarshal(w.Body.Bytes(), &resp))
	data, ok := resp["data"].(map[string]any)
	require.True(t, ok, "response should have 'data' key")
	assert.NotEmpty(t, data["access_token"])
	assert.Greater(t, data["expires_in"], float64(0))

	cookie := findCookie(w, "refresh_token")
	require.NotNil(t, cookie, "new refresh_token cookie should be set")
	assert.NotEmpty(t, cookie.Value)
	assert.NotEqual(t, "old-raw-token", cookie.Value)
	assert.True(t, cookie.HttpOnly)
}

func TestHandler_Refresh_NoCookie(t *testing.T) {
	t.Parallel()
	r, _, _ := newHandlerRouter(t)

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "MISSING_TOKEN")
}

func TestHandler_Refresh_InvalidToken(t *testing.T) {
	t.Parallel()
	r, _, authRepo := newHandlerRouter(t)

	authRepo.getByHashFn = func(_ context.Context, _ string) (*auth.RefreshToken, error) {
		return nil, auth.ErrRefreshTokenNotFound
	}

	w := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/api/v1/auth/refresh", nil)
	req.AddCookie(&http.Cookie{Name: "refresh_token", Value: "bogus-token"})
	r.ServeHTTP(w, req)

	assert.Equal(t, http.StatusUnauthorized, w.Code)
	assert.Contains(t, w.Body.String(), "INVALID_TOKEN")
}
