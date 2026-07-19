package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
)

type App struct {
	Logger *logger.Logger
	Config *config.Config
	Router *gin.Engine
}

func NewApp() (*App, error) {
	cfg, err := config.Load(config.LoaderOptions{
		ConfigFile: "configs/config.yaml",
	})
	if err != nil {
		return nil, err
	}

	// Initialize the logger
	root, err := logger.New(cfg.Logging)
	if err != nil {
		return nil, err
	}

	// Setup router and middleware
	router := SetupRouter(cfg, *root.Zerolog)

	return &App{
		Logger: root,
		Config: cfg,
		Router: router,
	}, nil
}

func CloseApp(app *App) error {
	if err := app.Logger.Close(); err != nil {
		return err
	}

	return nil
}
