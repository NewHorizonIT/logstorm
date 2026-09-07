package project

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, h *ProjectHandler) {
	g := router.Group("/projects")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.GET("/:id", h.Get)
	g.PUT("/:id", h.Update)
}
