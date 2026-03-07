package ingestion

import (
	"github.com/NewHorizonIT/logstorm/internal/services/event"
	"github.com/gin-gonic/gin"
)

func SetupIngestionRoutes(r *gin.RouterGroup, publisher event.IPublisher) {
	ingestionHanlder := NewIngestionHandler(publisher)
	r.POST("/logs", ingestionHanlder.IngestLogs)
}
