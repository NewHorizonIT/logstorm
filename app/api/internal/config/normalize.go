package config

import "strings"

func Normalize(cfg *Config) {
	// lowercase the logging level for consistency
	cfg.Logging.Level = strings.ToLower(cfg.Logging.Level)

	// Lowercase the protocol for consistency
	cfg.Server.Protocol = strings.ToLower(cfg.Server.Protocol)

	// Standardize the base path to always start with a single slash and not end with a slash
	if !strings.HasPrefix(cfg.Server.BasePath, "/") {
		cfg.Server.BasePath = "/" + cfg.Server.BasePath
	}
	if strings.HasSuffix(cfg.Server.BasePath, "/") {
		cfg.Server.BasePath = strings.TrimSuffix(cfg.Server.BasePath, "/")
	}

}
