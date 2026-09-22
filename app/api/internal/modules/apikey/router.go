package apikey

import "github.com/gin-gonic/gin"

func RegisterRoutes(rg *gin.RouterGroup, h *APIKeyHandler) {
	g := rg.Group("/api-keys")
	g.POST("", h.Create)
	g.GET("", h.List)
	g.DELETE("/:id", h.Revoke)
}
