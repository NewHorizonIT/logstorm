package middleware

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/stretchr/testify/assert"
)

func TestCORS_AllowsConfiguredOrigin(t *testing.T) {
	gin.SetMode(gin.TestMode)

	cfg := config.CORSConfig{
		AllowedOrigins: []string{
			"http://localhost:3000",
		},
		AllowedMethods: []string{
			"GET",
			"POST",
			"PUT",
			"DELETE",
		},
		AllowedHeaders: []string{
			"Authorization",
			"Content-Type",
		},
	}

	router := gin.New()

	router.Use(
		CORS(cfg),
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

	req.Header.Set(
		"Origin",
		"http://localhost:3000",
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

	assert.Equal(
		t,
		"http://localhost:3000",
		recorder.Header().Get(
			"Access-Control-Allow-Origin",
		),
	)

	assert.Equal(
		t,
		"GET, POST, PUT, DELETE",
		recorder.Header().Get(
			"Access-Control-Allow-Methods",
		),
	)

	assert.Equal(
		t,
		"Authorization, Content-Type",
		recorder.Header().Get(
			"Access-Control-Allow-Headers",
		),
	)

	assert.Equal(
		t,
		"Origin",
		recorder.Header().Get("Vary"),
	)
}

func TestIsAllowedOrigin(t *testing.T) {
	tests := []struct {
		name           string
		origin         string
		allowedOrigins []string
		expected       bool
	}{
		{
			name:   "allowed origin",
			origin: "http://localhost:3000",
			allowedOrigins: []string{
				"http://localhost:3000",
			},
			expected: true,
		},
		{
			name:   "unknown origin",
			origin: "http://evil.com",
			allowedOrigins: []string{
				"http://localhost:3000",
			},
			expected: false,
		},
		{
			name:   "empty origin",
			origin: "",
			allowedOrigins: []string{
				"http://localhost:3000",
			},
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isAllowedOrigin(
				tt.origin,
				tt.allowedOrigins,
			)

			assert.Equal(
				t,
				tt.expected,
				result,
			)
		})
	}
}
