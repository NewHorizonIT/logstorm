package dashboard

import "github.com/gin-gonic/gin"

func RegisterRoutes(r *gin.RouterGroup, h *DashboardHandler) {
	g := r.Group("/dashboard")
	g.GET("/:project_id", h.GetStats)
}
