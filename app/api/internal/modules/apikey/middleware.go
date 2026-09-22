package apikey

import (
	"crypto/sha256"
	"encoding/hex"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/response"
)

const ProjectIDKey = "apikey_project_id"

type Middleware struct {
	ValidateAPIKey gin.HandlerFunc
}

func NewMiddleware(repo APIKeyRepository) Middleware {
	return Middleware{
		ValidateAPIKey: func(c *gin.Context) {
			rawKey := c.GetHeader("X-API-Key")
			if rawKey == "" {
				response.AbortUnauthorized(c, "MISSING_API_KEY", "X-API-Key header is required")
				return
			}

			sum := sha256.Sum256([]byte(rawKey))
			keyHash := hex.EncodeToString(sum[:])

			key, err := repo.GetByHash(c.Request.Context(), keyHash)
			if err != nil {
				response.AbortUnauthorized(c, "INVALID_API_KEY", "API key is invalid or revoked")
				return
			}

			c.Set(ProjectIDKey, key.ProjectID)
			c.Next()
		},
	}
}

// ProjectIDFromContext lấy project_id được inject bởi ValidateAPIKey middleware.
func ProjectIDFromContext(c *gin.Context) (uuid.UUID, bool) {
	v, exists := c.Get(ProjectIDKey)
	if !exists {
		return uuid.Nil, false
	}
	id, ok := v.(uuid.UUID)
	return id, ok
}
