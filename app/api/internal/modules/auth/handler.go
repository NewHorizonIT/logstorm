package auth

import (
	"errors"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/user"
	"github.com/logstorm/api/internal/response"
)

const refreshTokenCookie = "refresh_token"

type AuthHandler struct {
	authSvc  *AuthService
	tokenSvc *TokenService
	secure   bool
}

func NewAuthHandler(authSvc *AuthService, tokenSvc *TokenService, cfg *config.Config) *AuthHandler {
	return &AuthHandler{
		authSvc:  authSvc,
		tokenSvc: tokenSvc,
		secure:   cfg.App.Environment != "development",
	}
}

type registerRequest struct {
	Email    string `json:"email"     binding:"required"`
	FullName string `json:"full_name"  binding:"required"`
	Password string `json:"password"   binding:"required,min=8"`
}

type loginRequest struct {
	Email    string `json:"email"    binding:"required"`
	Password string `json:"password" binding:"required,min=8"`
}

type tokenResponse struct {
	AccessToken string `json:"access_token"`
	TokenType   string `json:"token_type"`
	ExpiresIn   int    `json:"expires_in"`
}

type userResponse struct {
	ID         string `json:"id"`
	Email      string `json:"email"`
	FullName   string `json:"full_name"`
	Status     string `json:"status"`
	IsVerified bool   `json:"is_verified"`
	CreatedAt  string `json:"created_at"`
}

func (h *AuthHandler) Register(c *gin.Context) {
	var req registerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	u, err := h.authSvc.Register(c.Request.Context(), RegisterInput{
		Email:    req.Email,
		FullName: req.FullName,
		Password: req.Password,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, toUserResponse(u))
}

func (h *AuthHandler) Login(c *gin.Context) {
	var req loginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	accessToken, rawRefresh, err := h.authSvc.Login(c.Request.Context(), LoginInput{
		Email:    req.Email,
		Password: req.Password,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshCookie(c, rawRefresh)
	response.OK(c, tokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.tokenSvc.AccessTokenTTL().Seconds()),
	})
}

func (h *AuthHandler) Logout(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshTokenCookie)
	if err == nil {
		if logoutErr := h.authSvc.Logout(c.Request.Context(), rawRefresh); logoutErr != nil {
			log := logger.FromContext(c.Request.Context())
			log.Error().Err(logoutErr).Msg("auth: logout failed to revoke token")
		}
	}
	h.clearRefreshCookie(c)
	response.NoContent(c)
}

func (h *AuthHandler) Refresh(c *gin.Context) {
	rawRefresh, err := c.Cookie(refreshTokenCookie)
	if err != nil {
		response.Unauthorized(c, "MISSING_TOKEN", "refresh token is required")
		return
	}

	accessToken, newRawRefresh, err := h.authSvc.Refresh(c.Request.Context(), rawRefresh)
	if err != nil {
		h.handleError(c, err)
		return
	}

	h.setRefreshCookie(c, newRawRefresh)
	response.OK(c, tokenResponse{
		AccessToken: accessToken,
		TokenType:   "Bearer",
		ExpiresIn:   int(h.tokenSvc.AccessTokenTTL().Seconds()),
	})
}

func (h *AuthHandler) setRefreshCookie(c *gin.Context, rawToken string) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    rawToken,
		MaxAge:   int(h.tokenSvc.RefreshTokenTTL().Seconds()),
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) clearRefreshCookie(c *gin.Context) {
	http.SetCookie(c.Writer, &http.Cookie{
		Name:     refreshTokenCookie,
		Value:    "",
		MaxAge:   -1,
		Path:     "/",
		HttpOnly: true,
		Secure:   h.secure,
		SameSite: http.SameSiteStrictMode,
	})
}

func (h *AuthHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrInvalidCredentials):
		response.Unauthorized(c, "INVALID_CREDENTIALS", "invalid email or password")
	case errors.Is(err, ErrRefreshTokenNotFound):
		response.Unauthorized(c, "INVALID_TOKEN", "refresh token is invalid or expired")
	case errors.Is(err, ErrTokenExpired):
		response.Unauthorized(c, "TOKEN_EXPIRED", "refresh token has expired")
	case errors.Is(err, user.ErrUserEmailAlreadyExists):
		response.Conflict(c, "EMAIL_ALREADY_EXISTS", "a user with this email already exists")
	case errors.Is(err, user.ErrValidation):
		response.UnprocessableEntity(c, "VALIDATION_ERROR", err.Error())
	default:
		log := logger.FromContext(c.Request.Context())
		log.Error().Err(err).Msg("auth: unexpected handler error")
		response.InternalServerError(c)
	}
}

func toUserResponse(u *user.User) userResponse {
	return userResponse{
		ID:         u.ID.String(),
		Email:      u.Email,
		FullName:   u.FullName,
		Status:     u.Status,
		IsVerified: u.IsVerified,
		CreatedAt:  u.CreatedAt.Format(time.RFC3339),
	}
}
