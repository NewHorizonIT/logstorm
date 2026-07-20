package bootstrap

import (
	"github.com/gin-gonic/gin"
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/database"
	"github.com/logstorm/api/internal/logger"
)

type App struct {
	Logger *logger.Logger
	Config *config.Config
	Router *gin.Engine
	DB     *database.Postgres
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

	// Connect to the database
	postgres, err := database.Connect(cfg.Database)
	if err != nil {
		return nil, err
	}

	// Ping the database to ensure the connection is valid
	if err := postgres.Ping(); err != nil {
		return nil, err
	}
	// Setup router and middleware
	router := SetupRouter(cfg, *root.Zerolog)

	return &App{
		Logger: root,
		Config: cfg,
		Router: router,
		DB:     postgres,
	}, nil
}

func CloseApp(app *App) error {
	if err := app.Logger.Close(); err != nil {
		return err
	}

	app.DB.Close()

	return nil
}
