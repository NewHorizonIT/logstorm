package middleware

import (
	"time"

	"github.com/NewHorizonIT/logstorm-ingestion/internal/observability"
	"github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		c.Next()

		duration := time.Since(start).Seconds()
		observability.HTTPRequestDuration.Observe(duration)
		observability.HTTPRequestsTotal.WithLabelValues(c.Request.Method, "200").Inc()
	}
}
