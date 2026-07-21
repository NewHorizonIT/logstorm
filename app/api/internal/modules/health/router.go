package health

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

func RegisterRoutes(router *gin.RouterGroup) {
	router.GET("/health", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"status":  "ok",
				"service": "LogStorm API",
				"version": "1.0.0",
			},
		)
	})
}
