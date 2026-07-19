package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/middleware"
	"github.com/logstorm/api/internal/modules/health"
	"github.com/rs/zerolog"
)

func SetupRouter(
	cfg *config.Config,
	appLogger zerolog.Logger,
) *gin.Engine {
	router := gin.New()

	api := router.Group(cfg.Server.BasePath)

	api.Use(
		middleware.Recovery(appLogger),
		logger.RequestLogger(appLogger),
		middleware.CORS(cfg.CORS),
	)

	// Add your routes here
	health.RegisterRoutes(api)

	return router
}
