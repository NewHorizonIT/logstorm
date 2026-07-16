package middleware

import (
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
)

const (
	HeaderOrigin              = "Origin"
	AccessControlAllowOrigin  = "Access-Control-Allow-Origin"
	Vary                      = "Vary"
	Origin                    = "Origin"
	AccessControlAllowMethods = "Access-Control-Allow-Methods"
	AccessControlAllowHeaders = "Access-Control-Allow-Headers"
)

func CORS(cfg config.CORSConfig) gin.HandlerFunc {
	return func(c *gin.Context) {
		origin := c.GetHeader(HeaderOrigin)

		if origin != "" && isAllowedOrigin(origin, cfg.AllowedOrigins) {
			c.Header(AccessControlAllowOrigin, origin)
			c.Header(Vary, Origin)
		}

		c.Header(AccessControlAllowMethods, strings.Join(cfg.AllowedMethods, ", "))

		c.Header(
			AccessControlAllowHeaders,
			strings.Join(cfg.AllowedHeaders, ", "),
		)

		if c.Request.Method == http.MethodOptions {
			c.AbortWithStatus(http.StatusNoContent)
			return
		}

		c.Next()
	}
}

func isAllowedOrigin(origin string, allowedOrigins []string) bool {
	for _, allowedOrigin := range allowedOrigins {
		if origin == allowedOrigin {
			return true
		}
	}

	return false
}
