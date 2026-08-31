package auth

import "github.com/gin-gonic/gin"

func RegisterRoutes(router *gin.RouterGroup, h *AuthHandler) {
	g := router.Group("/auth")
	g.POST("/register", h.Register)
	g.POST("/login", h.Login)
	g.POST("/logout", h.Logout)
	g.POST("/refresh", h.Refresh)
}
