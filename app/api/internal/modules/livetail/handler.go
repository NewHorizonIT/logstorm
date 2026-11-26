package livetail

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"

	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
	"github.com/logstorm/api/internal/response"
)

const heartbeatInterval = 30 * time.Second

type Subscriber interface {
	Subscribe(projectID uuid.UUID, filter Filter) (<-chan *search.LogRecord, func())
}

type Handler struct {
	hub         Subscriber
	projectRepo project.ProjectRepository
}

func NewHandler(hub *Hub, projectRepo project.ProjectRepository) *Handler {
	return &Handler{hub: hub, projectRepo: projectRepo}
}

func (h *Handler) Stream(c *gin.Context) {
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

	if _, err := h.projectRepo.GetByID(c.Request.Context(), projectID, ownerID); err != nil {
		if errors.Is(err, project.ErrProjectNotFound) {
			response.Forbidden(c, "FORBIDDEN", "access denied")
			return
		}
		response.InternalServerError(c)
		return
	}

	filter := Filter{
		Level:   c.Query("level"),
		Service: c.Query("service"),
	}

	ch, unsub := h.hub.Subscribe(projectID, filter)
	defer unsub()

	c.Writer.Header().Set("Content-Type", "text/event-stream")
	c.Writer.Header().Set("Cache-Control", "no-cache")
	c.Writer.Header().Set("Connection", "keep-alive")
	c.Writer.Header().Set("X-Accel-Buffering", "no")
	c.Writer.WriteHeader(http.StatusOK)
	c.Writer.Flush()

	ticker := time.NewTicker(heartbeatInterval)
	defer ticker.Stop()

	ctx := c.Request.Context()
	for {
		select {
		case record := <-ch:
			data, err := json.Marshal(record)
			if err != nil {
				continue
			}
			fmt.Fprintf(c.Writer, "event: log\ndata: %s\n\n", data)
			c.Writer.Flush()

		case <-ticker.C:
			fmt.Fprintf(c.Writer, ": heartbeat\n\n")
			c.Writer.Flush()

		case <-ctx.Done():
			return
		}
	}
}
