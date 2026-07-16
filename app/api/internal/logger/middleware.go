package logger

import (
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog"
)

func RequestLogger(appLogger zerolog.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()

		requestID := c.GetHeader(
			"X-Request-ID",
		)

		if requestID == "" {
			requestID = uuid.NewString()
		}

		c.Header(
			"X-Request-ID",
			requestID,
		)

		requestLogger := appLogger.With().
			Str(FieldRequestID, requestID).
			Str(FieldMethod, c.Request.Method).
			Str(FieldPath, c.Request.URL.Path).
			Logger()

		ctx := WithLogger(
			c.Request.Context(),
			requestLogger,
		)

		c.Request = c.Request.WithContext(ctx)

		c.Next()

		duration := time.Since(start)

		logHTTPResponse(
			requestLogger,
			c,
			duration,
		)
	}
}

func logHTTPResponse(log zerolog.Logger, c *gin.Context, duration time.Duration) {
	status := c.Writer.Status()
	event := log.Info()

	switch {
	case status >= 500:
		event = log.Error()

	case status >= 400:
		event = log.Warn()
	}

	event.Int(FieldStatus, status).
		Int64(FieldDurationMS, duration.Milliseconds()).
		Str("client_ip", c.ClientIP()).
		Str("user_agent", c.Request.UserAgent()).
		Msg("http_request_completed")
}
