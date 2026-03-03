package handler

import (
	"fmt"
	"net/http"
	"time"

	"github.com/NewHorizonIT/logstorm-ingestion/internal/middleware"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/model"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/producer"
	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type IngestHandler struct {
	producer *producer.KafkaProducer
	limiter  *middleware.Manager
}

func NewIngestHandler(p *producer.KafkaProducer, limiter *middleware.Manager) *IngestHandler {
	return &IngestHandler{producer: p, limiter: limiter}
}

func (h *IngestHandler) HandleIngest(c *gin.Context) {
	var req model.LogEvent

	if err := c.ShouldBindJSON(&req); err != nil {
		fmt.Printf("Invalid request: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if !h.limiter.Allow(req.Service) {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "rate limit exceeded"})
		return
	}

	// Enrich metadata
	req.ID = uuid.New().String()
	req.Timestamp = time.Now()
	req.Env = "dev"

	if err := h.producer.Enqueue(req); err != nil {
		c.JSON(http.StatusTooManyRequests, gin.H{"error": "buffer full"})
		return
	}

	c.Status(http.StatusAccepted)
}

func (h *IngestHandler) MetricsHandler() gin.HandlerFunc {
	return gin.WrapH(promhttp.Handler())
}
