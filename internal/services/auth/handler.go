package auth

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

type AuthHandler struct {
	usecase AuthUsecase
}

func NewAuthHandler(usecase AuthUsecase) *AuthHandler {
	return &AuthHandler{usecase: usecase}
}

func (ah *AuthHandler) LoginHandler(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	token, err := ah.usecase.Login(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, LoginResponse{AccessToken: token})
}

func (ah *AuthHandler) RegisterHandler(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var result *RegisterResult
	result, err := ah.usecase.Register(c.Request.Context(), req.Email, req.Password)
	if err != nil {
		c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
		return
	}

	// Refresh token cookie HTTP-only
	c.SetCookie("refresh_token", result.RefreshToken, 7*24*3600, "/", "", false, true)

	res := &RegisterResponse{
		Account:     AccountDTO{ID: result.Account.ID, Email: result.Account.Email},
		AccessToken: result.AccessToken,
	}
	c.JSON(http.StatusCreated, res)
}

func (ah *AuthHandler) RefreshHandler(c *gin.Context) {
	// Step 1: Get token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "refresh token required"})
		return
	}
	// Step 2: call usecase to refresh tokens
	result, err := ah.usecase.Refresh(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
		return
	}

	// Step 3: set new refresh token cookie
	c.SetCookie("refresh_token", result.RefreshToken, 7*24*3600, "/", "", false, true)

	// Step 4: return new access token
	res := HandleRefreshTokenResponse{AccessToken: result.AccessToken}
	c.JSON(http.StatusOK, res)
}

func (ah *AuthHandler) LogoutHandler(c *gin.Context) {
	// Step 1: Get token from cookie
	refreshToken, err := c.Cookie("refresh_token")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "refresh token required"})
		return
	}

	// Step 2: call usecase to logout
	err = ah.usecase.Logout(c.Request.Context(), refreshToken)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	// Step 3: clear refresh token cookie
	c.SetCookie("refresh_token", "", -1, "/", "", false, true)

	c.JSON(http.StatusOK, gin.H{"message": "logged out successfully"})
}
