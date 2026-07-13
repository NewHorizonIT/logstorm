package bootstrap

import "github.com/logstorm/api/internal/config"

type App struct {
	Config *config.Config
}

func NewApp() *App {
	cfg, err := config.Load(config.LoaderOptions{
		ConfigFile: "../../configs/config.yaml",
	})
	if err != nil {
		panic(err)
	}

	return &App{
		Config: cfg,
	}
}
