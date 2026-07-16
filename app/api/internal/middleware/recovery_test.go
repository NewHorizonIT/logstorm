package middleware

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func TestRecovery_RecoversFromPanic(t *testing.T) {
	gin.SetMode(gin.TestMode)

	var logBuffer bytes.Buffer

	logger := zerolog.New(&logBuffer)

	router := gin.New()

	router.Use(
		Recovery(logger),
	)

	router.GET("/panic", func(c *gin.Context) {
		panic("something went wrong")
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/panic",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusInternalServerError,
		recorder.Code,
	)

	assert.JSONEq(
		t,
		`{
			"code": "INTERNAL_SERVER_ERROR",
			"message": "internal server error"
		}`,
		recorder.Body.String(),
	)

	assert.Contains(
		t,
		logBuffer.String(),
		"panic_recovered",
	)

	assert.Contains(
		t,
		logBuffer.String(),
		"something went wrong",
	)
}

func TestRecovery_DoesNotAffectNormalRequest(t *testing.T) {
	gin.SetMode(gin.TestMode)

	logger := zerolog.Nop()

	router := gin.New()

	router.Use(
		Recovery(logger),
	)

	router.GET("/health", func(c *gin.Context) {
		c.JSON(
			http.StatusOK,
			gin.H{
				"status": "ok",
			},
		)
	})

	req := httptest.NewRequest(
		http.MethodGet,
		"/health",
		nil,
	)

	recorder := httptest.NewRecorder()

	router.ServeHTTP(
		recorder,
		req,
	)

	assert.Equal(
		t,
		http.StatusOK,
		recorder.Code,
	)

	assert.JSONEq(
		t,
		`{
			"status": "ok"
		}`,
		recorder.Body.String(),
	)
}
