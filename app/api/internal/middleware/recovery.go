package middleware

import (
	"net/http"
	"runtime/debug"

	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/logger"
	"github.com/rs/zerolog"
)

func Recovery(appLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		defer func() {
			if recovered := recover(); recovered != nil {
				appLogger.Error().
					Any(logger.FieldRecovered, recovered).
					Bytes(logger.FieldStack, debug.Stack()).
					Str(logger.FieldMethod, c.Request.Method).
					Str(logger.FieldPath, c.Request.URL.Path).
					Str(logger.FieldRequestID, c.GetHeader("X-Request-ID")).
					Msg("panic_recovered")

				c.AbortWithStatusJSON(
					http.StatusInternalServerError,
					gin.H{
						"code":    "INTERNAL_SERVER_ERROR",
						"message": "internal server error",
					},
				)
			}
		}()

		c.Next()
	}
}
