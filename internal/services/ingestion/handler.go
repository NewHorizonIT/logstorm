package ingestion

import (
	"github.com/NewHorizonIT/logstorm/internal/domain"
	"github.com/NewHorizonIT/logstorm/internal/services/event"
	"github.com/NewHorizonIT/logstorm/pkg"
	"github.com/gin-gonic/gin"
)

type IngestionHandler struct {
	publisher event.IPublisher
}

func NewIngestionHandler(publisher event.IPublisher) *IngestionHandler {
	return &IngestionHandler{
		publisher: publisher,
	}
}

func (h *IngestionHandler) IngestLogs(c *gin.Context) {
	ctx := c.Request.Context()

	var log domain.Log
	if err := c.ShouldBindJSON(&log); err != nil {
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	body, err := pkg.JsonToBytes(log)
	if err != nil {
		c.JSON(500, gin.H{"error": "failed to encode log"})
		return
	}

	// int64 marshal never fails; error intentionally ignored
	key, _ := pkg.JsonToBytes(log.ID)

	if err := h.publisher.Publish(ctx, "logs-topic", key, body); err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}

	c.JSON(200, gin.H{"message": "Log ingested successfully"})
}
