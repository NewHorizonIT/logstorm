package search

import "github.com/gin-gonic/gin"

func RegisterSearchRoutes(r *gin.RouterGroup, h *SearchHandler) {
	searchRouter := r.Group("/search")

	searchRouter.GET("/:project_id/logs", h.SearchLogs)
}
