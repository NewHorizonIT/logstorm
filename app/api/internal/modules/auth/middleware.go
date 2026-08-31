package auth

import (
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/response"
)

const UserIDKey = "auth_user_id"

// Middleware holds Gin middleware functions for the auth domain.
type Middleware struct {
	Authenticate gin.HandlerFunc
}

func NewMiddleware(tokenSvc *TokenService) Middleware {
	return Middleware{
		Authenticate: func(c *gin.Context) {
			header := c.GetHeader("Authorization")
			if !strings.HasPrefix(header, "Bearer ") {
				response.AbortUnauthorized(c, "MISSING_TOKEN", "authorization header is required")
				return
			}

			claims, err := tokenSvc.ValidateAccessToken(strings.TrimPrefix(header, "Bearer "))
			if err != nil {
				response.AbortUnauthorized(c, "INVALID_TOKEN", "access token is invalid or expired")
				return
			}

			c.Set(UserIDKey, claims.UserID)
			c.Next()
		},
	}
}

// UserIDFromContext extracts the authenticated user ID set by Middleware.Authenticate.
func UserIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	id, exists := c.Get(UserIDKey)
	if !exists {
		return uuid.Nil, false
	}
	uid, ok := id.(uuid.UUID)
	return uid, ok
}
