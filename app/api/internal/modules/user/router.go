package user

import "github.com/gin-gonic/gin"

func RegisterRoutes(group *gin.RouterGroup, h *UserHandler) {
	group.GET("/users/me", h.Me)
}
