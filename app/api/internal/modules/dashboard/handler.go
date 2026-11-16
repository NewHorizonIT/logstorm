package dashboard

import (
	"errors"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/response"
)

type DashboardHandler struct {
	svc *DashboardService
}

func NewDashboardHandler(svc *DashboardService) *DashboardHandler {
	return &DashboardHandler{svc: svc}
}

func (h *DashboardHandler) GetStats(c *gin.Context) {
	ownerID, ok := auth.UserIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_TOKEN", "authentication required")
		return
	}

	projectID, err := uuid.Parse(c.Param("project_id"))
	if err != nil {
		response.BadRequest(c, "INVALID_PROJECT_ID", "project id must be a valid UUID")
		return
	}

	var from, to time.Time

	if fromStr := c.Query("from"); fromStr != "" {
		t, err := time.Parse(time.RFC3339, fromStr)
		if err != nil {
			response.BadRequest(c, "INVALID_FROM_TIME", "from must be RFC3339 format")
			return
		}
		from = t
	}

	if toStr := c.Query("to"); toStr != "" {
		t, err := time.Parse(time.RFC3339, toStr)
		if err != nil {
			response.BadRequest(c, "INVALID_TO_TIME", "to must be RFC3339 format")
			return
		}
		to = t
	}

	stats, err := h.svc.GetStats(c.Request.Context(), ownerID, projectID, from, to)
	if err != nil {
		switch {
		case errors.Is(err, ErrForbidden):
			response.Forbidden(c, "FORBIDDEN", "access denied")
		case errors.Is(err, ErrInvalidTimeRange):
			response.BadRequest(c, "INVALID_TIME_RANGE", "invalid input")
		case errors.Is(err, ErrTimeRangeTooLarge):
			response.BadRequest(c, "TIME_RANGE_TOO_LARGE", "time range too large")
		default:
			log := logger.FromContext(c.Request.Context())
			log.Error().Err(err).Msg("dashboard: unexpected handler error")
			response.InternalServerError(c)
		}
		return
	}

	response.OK(c, stats)
}
