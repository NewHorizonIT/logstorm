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
}

func (ah *AuthHandler) RefreshHandler(c *gin.Context) {
}

func (ah *AuthHandler) LogoutHandler(c *gin.Context) {
}
