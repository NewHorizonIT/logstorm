package config

import (
	"fmt"
	"strings"

	"github.com/spf13/viper"
)

type LoaderOptions struct {
	ConfigFile string
}

func Load(opts LoaderOptions) (*Config, error) {
	var config Config

	v := viper.New()

	// Reading YAML configuration file
	v.SetConfigFile(opts.ConfigFile)

	// Default values
	v.SetDefault("app.name", "LogStorm")
	v.SetDefault("server.protocol", "http")
	v.SetDefault("server.host", "localhost")
	v.SetDefault("server.port", 8080)
	v.SetDefault("server.base_path", "/api/v1")
	v.SetDefault("logging.level", "info")
	v.SetDefault("logging.format", "json")
	v.SetDefault("logging.output", "stdout")
	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("error reading config file: %w", err)
	}

	// Read .env file if it exists
	v.AutomaticEnv()
	v.SetEnvKeyReplacer(
		strings.NewReplacer(".", "_"),
	)

	if err := v.Unmarshal(&config); err != nil {
		return nil, fmt.Errorf("error unmarshaling config: %w", err)
	}

	// Validate the config struct
	if err := Validate(&config); err != nil {
		return nil, fmt.Errorf("error validating config: %w", err)
	}

	// Normalize the config struct
	Normalize(&config)

	return &config, nil
}
