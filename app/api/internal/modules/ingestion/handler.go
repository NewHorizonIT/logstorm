package ingestion

import (
	"context"
	"errors"

	"github.com/gin-gonic/gin"

	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/logstorm/api/internal/response"
)

// Publisher is the port the handler uses to push events into the pipeline.
type Publisher interface {
	Publish(ctx context.Context, event *LogEvent) error
}

type IngestionHandler struct {
	publisher Publisher
}

func NewIngestionHandler(publisher Publisher) *IngestionHandler {
	return &IngestionHandler{publisher: publisher}
}

func (h *IngestionHandler) Ingest(c *gin.Context) {
	projectID, ok := apikey.ProjectIDFromContext(c)
	if !ok {
		response.Unauthorized(c, "MISSING_API_KEY", "authentication required")
		return
	}

	var input IngestLogInput
	if err := c.ShouldBindJSON(&input); err != nil {
		response.BadRequest(c, "INVALID_REQUEST", err.Error())
		return
	}

	event, err := input.ToLogEvent(projectID, "sdk")
	if err != nil {
		if errors.Is(err, ErrValidation) {
			response.UnprocessableEntity(c, "VALIDATION_ERROR", err.Error())
			return
		}
		log := logger.FromContext(c.Request.Context())
		log.Error().Err(err).Msg("ingestion: unexpected error building log event")
		response.InternalServerError(c)
		return
	}

	if err := h.publisher.Publish(c.Request.Context(), event); err != nil {
		log := logger.FromContext(c.Request.Context())
		log.Error().Err(err).Msg("ingestion: publish failed")
		response.InternalServerError(c)
		return
	}

	response.Accepted(c, gin.H{"id": event.ID})
}
