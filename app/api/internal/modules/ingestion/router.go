package ingestion

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *IngestionHandler) {
	rg.POST("/logs", h.Ingest)
}
