package config

import (
	"fmt"

	"github.com/go-playground/validator/v10"
)

var validate = validator.New()

func Validate(cfg *Config) error {
	if err := validate.Struct(cfg); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// Cross Validation
	// MaxIdleConns <= MaxOpenConns
	if cfg.Database.MaxIdleConns > cfg.Database.MaxOpenConns {
		return fmt.Errorf("config validation failed: MaxIdleConns (%d) cannot be greater than MaxOpenConns (%d)",
			cfg.Database.MaxIdleConns, cfg.Database.MaxOpenConns)
	}

	// AccessTokenTTL < RefreshTokenTTL
	if cfg.Auth.AccessTokenTTL >= cfg.Auth.RefreshTokenTTL {
		return fmt.Errorf("config validation failed: AccessTokenTTL (%s) must be less than RefreshTokenTTL (%s)",
			cfg.Auth.AccessTokenTTL, cfg.Auth.RefreshTokenTTL)
	}

	return nil
}
