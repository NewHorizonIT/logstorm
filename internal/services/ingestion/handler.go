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
	// Parse the incoming log data
	var log domain.Log
	err := c.ShouldBindJSON(&log)

	// If error pushlih into dql-topic and return error response
	if err != nil {
		key := pkg.JsonToBytes(log.ID)
		h.publisher.Publish(ctx, "dql-topic", key, []byte(log.Message))
		c.JSON(400, gin.H{"error": err.Error()})
		return
	}

	// If no error pushlih into log-topic and return success response
	key := pkg.JsonToBytes(log.ID)
	err = h.publisher.Publish(ctx, "logs-topic", key, pkg.JsonToBytes(log))
	if err != nil {
		c.JSON(500, gin.H{"error": err.Error()})
		return
	}
	c.JSON(200, gin.H{"message": "Log ingested successfully"})
}
