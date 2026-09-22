package apikey

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/response"
)

type APIKeyHandler struct {
	svc *APIKeyService
}

func NewAPIKeyHandler(svc *APIKeyService) *APIKeyHandler {
	return &APIKeyHandler{svc: svc}
}

type createAPIKeyRequest struct {
	ProjectID string `json:"project_id" binding:"required"`
	Name      string `json:"name"       binding:"required"`
}

type apiKeyResponse struct {
	ID        string  `json:"id"`
	ProjectID string  `json:"project_id"`
	Name      string  `json:"name"`
	KeyPrefix string  `json:"key_prefix"`
	CreatedAt string  `json:"created_at"`
	RevokedAt *string `json:"revoked_at"`
}

type createAPIKeyResponse struct {
	apiKeyResponse
	Key string `json:"key"`
}

func (h *APIKeyHandler) Create(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	var req createAPIKeyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	projectID, err := uuid.Parse(req.ProjectID)
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project_id must be a valid UUID")
		return
	}

	key, rawKey, err := h.svc.CreateAPIKey(c.Request.Context(), ownerID, CreateAPIKeyInput{
		ProjectID: projectID,
		Name:      req.Name,
	})
	if err != nil {
		h.handleError(c, err)
		return
	}

	response.Created(c, createAPIKeyResponse{
		apiKeyResponse: toAPIKeyResponse(key),
		Key:            rawKey,
	})
}

func (h *APIKeyHandler) List(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	var projectID *uuid.UUID
	if raw := c.Query("project_id"); raw != "" {
		id, err := uuid.Parse(raw)
		if err != nil {
			response.BadRequest(c, "INVALID_PROJECT_ID", "project_id must be a valid UUID")
			return
		}
		projectID = &id
	}

	keys, err := h.svc.ListAPIKeys(c.Request.Context(), ownerID, projectID)
	if err != nil {
		h.handleError(c, err)
		return
	}

	out := make([]apiKeyResponse, len(keys))
	for i, k := range keys {
		out[i] = toAPIKeyResponse(k)
	}
	response.OK(c, out)
}

func (h *APIKeyHandler) Revoke(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	id, err := uuid.Parse(c.Param("id"))
	if err != nil {
		response.BadRequest(c, "INVALID_KEY_ID", "key id must be a valid UUID")
		return
	}

	if err := h.svc.RevokeAPIKey(c.Request.Context(), id, ownerID); err != nil {
		h.handleError(c, err)
		return
	}

	response.NoContent(c)
}

func (h *APIKeyHandler) handleError(c *gin.Context, err error) {
	switch {
	case errors.Is(err, ErrValidation):
		response.UnprocessableEntity(c, "VALIDATION_ERROR", err.Error())
	case errors.Is(err, ErrKeyNotFound):
		response.NotFound(c, "KEY_NOT_FOUND", "api key not found")
	case errors.Is(err, ErrProjectNotFound):
		response.NotFound(c, "PROJECT_NOT_FOUND", "project not found")
	default:
		log := logger.FromContext(c.Request.Context())
		log.Error().Err(err).Msg("apikey: unexpected handler error")
		response.InternalServerError(c)
	}
}

func toAPIKeyResponse(k *APIKey) apiKeyResponse {
	r := apiKeyResponse{
		ID:        k.ID.String(),
		ProjectID: k.ProjectID.String(),
		Name:      k.Name,
		KeyPrefix: k.KeyPrefix,
		CreatedAt: k.CreatedAt.Format(time.RFC3339),
	}
	if k.RevokedAt != nil {
		s := k.RevokedAt.Format(time.RFC3339)
		r.RevokedAt = &s
	}
	return r
}
