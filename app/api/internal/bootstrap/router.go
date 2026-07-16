package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/middleware"
	"github.com/rs/zerolog"
)

func SetupRouter(
	cfg *config.Config,
	appLogger zerolog.Logger,
) *gin.Engine {
	router := gin.New()

	router.Use(
		logger.RequestLogger(appLogger),
		middleware.Recovery(appLogger),
		middleware.CORS(cfg.CORS),
	)

	return router
}
