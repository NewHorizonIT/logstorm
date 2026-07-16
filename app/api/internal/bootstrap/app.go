package bootstrap

import (
	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
)

type App struct {
	Logger *logger.Logger
	Config *config.Config
}

func NewApp() (*App, error) {
	cfg, err := config.Load(config.LoaderOptions{
		ConfigFile: "../../configs/config.yaml",
	})
	if err != nil {
		return nil, err
	}

	// Initialize the logger
	root, err := logger.New(cfg.Logging)
	if err != nil {
		return nil, err
	}

	return &App{
		Logger: root,
		Config: cfg,
	}, nil
}
