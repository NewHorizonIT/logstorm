package search

import (
	"errors"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/response"
)

type SearchHandler struct {
	svc SearchService
}

func NewSearchHandler(searchService *SearchService) *SearchHandler {
	return &SearchHandler{svc: *searchService}
}

func (h *SearchHandler) SearchLogs(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	projectID, err := parseUUID(c.Param("project_id"))
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project id must be a valid UUID")
		return
	}

	q := &SearchQuery{ProjectID: projectID}
	q.Level = c.Query("level")
	q.Service = c.Query("service")
	q.Text = c.Query("text")

	if fromStr := c.Query("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.BadRequest(c, "INVALID_FROM_TIME", "from must be RFC3339 format")
			return
		}
		q.From = t
	}

	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.BadRequest(c, "INVALID_TO_TIME", "to must be RFC3339 format")
			return
		}
		q.To = t
	}

	if limitStr := c.Query("limit"); limitStr != "" {
		n, err := strconv.Atoi(limitStr)
		if err != nil {
			response.BadRequest(c, "INVALID_LIMIT", "limit must be an integer")
			return
		}
		q.Limit = n
	}

	if offsetStr := c.Query("offset"); offsetStr != "" {
		n, err := strconv.Atoi(offsetStr)
		if err != nil {
			response.BadRequest(c, "INVALID_OFFSET", "offset must be an integer")
			return
		}
		q.Offset = n
	}

	result, err := h.svc.SearchLogs(c.Request.Context(), ownerID, q)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			response.Forbidden(c, "FORBIDDEN", "access denied")
		case errors.Is(err, ErrValidation):
			response.BadRequest(c, "VALIDATION_ERROR", err.Error())
		default:
			log := logger.FromContext(c.Request.Context())
			log.Error().Err(err).Msg("search: unexpected handler error")
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, gin.H{
		"logs":   result.Logs,
		"total":  result.Total,
		"limit":  q.Limit,
		"offset": q.Offset,
	})
}

func parseUUID(id string) (uuid.UUID, error) {
	return uuid.Parse(id)
}
