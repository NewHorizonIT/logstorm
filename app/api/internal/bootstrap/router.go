package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/middleware"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/health"
	"github.com/logstorm/api/internal/modules/project"
)

func SetupRouter(
	cfg *config.Config,
	log *logger.Logger,
	authHandler *auth.AuthHandler,
	authMiddleware auth.Middleware,
	projectHandler *project.ProjectHandler,
	apiKeyHandler *apikey.APIKeyHandler,
	apiKeyMiddleware apikey.Middleware,
) *gin.Engine {
	router := gin.New()

	api := router.Group(cfg.Server.BasePath)

	api.Use(
		middleware.Recovery(*log.Zerolog),
		logger.RequestLogger(*log.Zerolog),
		middleware.CORS(cfg.CORS),
	)

	health.RegisterRoutes(api)
	auth.RegisterRoutes(api, authHandler)

	protected := api.Group("", authMiddleware.Authenticate)
	project.RegisterRoutes(protected, projectHandler)
	apikey.RegisterRoutes(protected, apiKeyHandler)

	// SDK ingestion routes — authenticated by X-API-Key, not JWT
	ingestion := api.Group("", apiKeyMiddleware.ValidateAPIKey)
	_ = ingestion // log routes will be registered here in feat/logs

	return router
}
