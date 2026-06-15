package auth_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gin-gonic/gin"

	auth "github.com/NewHorizonIT/logstorm/internal/services/auth"
)

type mockUsecase struct {
	loginToken     string
	loginErr       error
	registerResult *auth.RegisterResult
	registerErr    error
	refreshResult  *auth.HandleRefreshTokenResult
	refreshErr     error
	logoutResult   *auth.LogoutResult
	logoutErr      error
}

func (m *mockUsecase) Login(ctx context.Context, email, password string) (string, error) {
	return m.loginToken, m.loginErr
}

func (m *mockUsecase) Register(ctx context.Context, email, password string) (*auth.RegisterResult, error) {
	return m.registerResult, m.registerErr
}

func (m *mockUsecase) Refresh(ctx context.Context, refreshToken string) (*auth.HandleRefreshTokenResult, error) {
	return m.refreshResult, m.refreshErr
}

func (m *mockUsecase) Logout(ctx context.Context, refreshToken string) error {
	return m.logoutErr
}

func TestLoginHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mu := &mockUsecase{loginToken: "tok123", loginErr: nil}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/login", h.LoginHandler)

		payload := map[string]string{"email": "u@example.com", "password": "hunter2"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf("expected status 200 got %d body=%s", w.Code, w.Body.String())
		}
		var resp auth.LoginResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.AccessToken != "tok123" {
			t.Fatalf("unexpected token: %s", resp.AccessToken)
		}
	})

	t.Run("bad_request", func(t *testing.T) {
		mu := &mockUsecase{}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/login", h.LoginHandler)

		// invalid JSON / missing required fields
		req := httptest.NewRequest(http.MethodPost, "/login", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("unauthorized", func(t *testing.T) {
		mu := &mockUsecase{loginErr: errors.New("invalid credentials")}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/login", h.LoginHandler)

		payload := map[string]string{"email": "u@example.com", "password": "wrongpass"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/login", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf("expected status 401 got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestRegisterHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success_sets_cookie_and_returns_account", func(t *testing.T) {
		rr := &auth.RegisterResult{
			Account:      &auth.Account{ID: 42, Email: "new@example.com"},
			AccessToken:  "access-xyz",
			RefreshToken: "refresh-xyz",
		}
		mu := &mockUsecase{registerResult: rr, registerErr: nil}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/register", h.RegisterHandler)

		payload := map[string]string{"email": "new@example.com", "password": "verysecure"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusCreated {
			t.Fatalf("expected status 201 got %d body=%s", w.Code, w.Body.String())
		}

		var resp auth.RegisterResponse
		if err := json.Unmarshal(w.Body.Bytes(), &resp); err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}
		if resp.Account.ID != 42 || resp.Account.Email != "new@example.com" {
			t.Fatalf("unexpected account in response: %+v", resp.Account)
		}
		if resp.AccessToken != "access-xyz" {
			t.Fatalf("unexpected access token: %s", resp.AccessToken)
		}

		// verify cookie
		res := w.Result()
		defer res.Body.Close()
		found := false
		for _, c := range res.Cookies() {
			if c.Name == "refresh_token" {
				found = true
				if c.Value != "refresh-xyz" {
					t.Fatalf("unexpected cookie value: %s", c.Value)
				}
				if !c.HttpOnly {
					t.Fatalf("cookie should be HttpOnly")
				}
			}
		}
		if !found {
			t.Fatalf("refresh_token cookie not set")
		}
	})

	t.Run("bad_request", func(t *testing.T) {
		mu := &mockUsecase{}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/register", h.RegisterHandler)

		// password too short (min=8) and missing email
		req := httptest.NewRequest(http.MethodPost, "/register", strings.NewReader("{}"))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusBadRequest {
			t.Fatalf("expected status 400 got %d body=%s", w.Code, w.Body.String())
		}
	})

	t.Run("conflict", func(t *testing.T) {
		mu := &mockUsecase{registerErr: errors.New("email already in use")}
		h := auth.NewAuthHandler(mu)
		r := gin.New()
		r.POST("/register", h.RegisterHandler)

		payload := map[string]string{"email": "taken@example.com", "password": "verysecure"}
		b, _ := json.Marshal(payload)
		req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewReader(b))
		req.Header.Set("Content-Type", "application/json")
		w := httptest.NewRecorder()
		r.ServeHTTP(w, req)

		if w.Code != http.StatusConflict {
			t.Fatalf("expected status 409 got %d body=%s", w.Code, w.Body.String())
		}
	})
}

func TestRefreshHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mu := &mockUsecase{
			refreshResult: &auth.HandleRefreshTokenResult{
				AccessToken:  "new-access-token",
				RefreshToken: "new-refresh-token",
			},
		}

		h := auth.NewAuthHandler(mu)

		r := gin.New()
		r.POST("/refresh", h.RefreshHandler)

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh",
			nil,
		)

		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "old-refresh-token",
		})

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200 got %d body=%s",
				w.Code,
				w.Body.String(),
			)
		}

		var resp auth.HandleRefreshTokenResponse

		err := json.Unmarshal(w.Body.Bytes(), &resp)
		if err != nil {
			t.Fatalf("unmarshal response: %v", err)
		}

		if resp.AccessToken != "new-access-token" {
			t.Fatalf(
				"expected access token %s got %s",
				"new-access-token",
				resp.AccessToken,
			)
		}

		found := false

		for _, c := range w.Result().Cookies() {
			if c.Name == "refresh_token" {
				found = true

				if c.Value != "new-refresh-token" {
					t.Fatalf(
						"expected cookie value %s got %s",
						"new-refresh-token",
						c.Value,
					)
				}

				if !c.HttpOnly {
					t.Fatal("cookie should be HttpOnly")
				}
			}
		}

		if !found {
			t.Fatal("refresh token cookie not found")
		}
	})

	t.Run("missing_cookie", func(t *testing.T) {
		mu := &mockUsecase{}

		h := auth.NewAuthHandler(mu)

		r := gin.New()

		r.POST("/refresh", h.RefreshHandler)

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh",
			nil,
		)

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status 401 got %d body=%s",
				w.Code,
				w.Body.String(),
			)
		}
	})
	t.Run("invalid_refresh_token", func(t *testing.T) {
		mu := &mockUsecase{
			refreshErr: errors.New("invalid refresh token"),
		}

		h := auth.NewAuthHandler(mu)

		r := gin.New()

		r.POST("/refresh", h.RefreshHandler)

		req := httptest.NewRequest(
			http.MethodPost,
			"/refresh",
			nil,
		)

		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "invalid-token",
		})

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusUnauthorized {
			t.Fatalf(
				"expected status 401 got %d body=%s",
				w.Code,
				w.Body.String(),
			)
		}
	})
}

func TestLogoutHandler(t *testing.T) {
	gin.SetMode(gin.TestMode)

	t.Run("success", func(t *testing.T) {
		mu := &mockUsecase{
			logoutErr: nil,
		}

		h := auth.NewAuthHandler(mu)

		r := gin.New()
		r.POST("/logout", h.LogoutHandler)

		req := httptest.NewRequest(
			http.MethodPost,
			"/logout",
			nil,
		)

		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "valid-refresh-token",
		})

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusOK {
			t.Fatalf(
				"expected status 200 got %d body=%s",
				w.Code,
				w.Body.String(),
			)
		}
	})

	t.Run("invalid_refresh_token", func(t *testing.T) {
		mu := &mockUsecase{
			logoutErr: errors.New("invalid refresh token"),
		}

		h := auth.NewAuthHandler(mu)

		r := gin.New()
		r.POST("/logout", h.LogoutHandler)

		req := httptest.NewRequest(
			http.MethodPost,
			"/logout",
			nil,
		)

		req.AddCookie(&http.Cookie{
			Name:  "refresh_token",
			Value: "invalid-refresh-token",
		})

		w := httptest.NewRecorder()

		r.ServeHTTP(w, req)

		if w.Code != http.StatusInternalServerError {
			t.Fatalf(
				"expected status 500 got %d body=%s",
				w.Code,
				w.Body.String(),
			)
		}
	})
}
