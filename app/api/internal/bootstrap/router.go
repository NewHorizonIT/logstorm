package bootstrap

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
	"github.com/logstorm/api/internal/middleware"
	"github.com/logstorm/api/internal/modules/apikey"
	"github.com/logstorm/api/internal/modules/auth"
	"github.com/logstorm/api/internal/modules/dashboard"
	"github.com/logstorm/api/internal/modules/health"
	"github.com/logstorm/api/internal/modules/ingestion"
	"github.com/logstorm/api/internal/modules/project"
	"github.com/logstorm/api/internal/modules/search"
)

func SetupRouter(
	cfg *config.Config,
	log *logger.Logger,
	authHandler *auth.AuthHandler,
	authMiddleware auth.Middleware,
	projectHandler *project.ProjectHandler,
	apiKeyHandler *apikey.APIKeyHandler,
	apiKeyMiddleware apikey.Middleware,
	ingestionHandler *ingestion.IngestionHandler,
	searchHandler *search.SearchHandler,
	dashboardHandler *dashboard.DashboardHandler,
) *gin.Engine {
	router := gin.New()

	// CORS at engine level so preflight OPTIONS requests are handled before route matching
	router.Use(middleware.CORS(cfg.CORS))
	router.OPTIONS("/*path", func(c *gin.Context) {
		c.Status(http.StatusNoContent)
	})

	api := router.Group(cfg.Server.BasePath)

	api.Use(
		middleware.Recovery(*log.Zerolog),
		logger.RequestLogger(*log.Zerolog),
	)

	health.RegisterRoutes(api)

	//set Allow-Credentials
	authGroup := api.Group("", middleware.AllowCredentials())
	auth.RegisterRoutes(authGroup, authHandler)

	protected := api.Group("", authMiddleware.Authenticate)
	project.RegisterRoutes(protected, projectHandler)
	apikey.RegisterRoutes(protected, apiKeyHandler)
	search.RegisterSearchRoutes(protected, searchHandler)
	dashboard.RegisterRoutes(protected, dashboardHandler)

	// SDK ingestion routes — authenticated by X-API-Key, not JWT
	sdkGroup := api.Group("", apiKeyMiddleware.ValidateAPIKey)
	ingestion.RegisterRoutes(sdkGroup, ingestionHandler)

	return router
}
