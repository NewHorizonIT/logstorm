package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLoadConfig_FromYAML(t *testing.T) {
	cfg, err := Load(LoaderOptions{
		ConfigFile: "../../configs/config_test.yaml",
	})
	require.NoError(t, err)
	assert.Equal(
		t,
		8080,
		cfg.Server.Port,
	)
	assert.Equal(
		t,
		"postgres",
		cfg.Database.Driver,
	)

}

func TestLoadConfig_ENVOverride(t *testing.T) {
	t.Setenv(
		"SERVER_PORT",
		"9999",
	)
	cfg, err := Load(LoaderOptions{
		ConfigFile: "../../configs/config_test.yaml",
	})
	require.NoError(t, err)
	assert.Equal(
		t,
		9999,
		cfg.Server.Port,
	)
}

func TestLoadConfig_DefaultValue(t *testing.T) {
	cfg, err := Load(LoaderOptions{
		ConfigFile: "../../configs/config_minimal.yaml",
	})
	require.NoError(t, err)
	assert.Equal(
		t,
		"info",
		cfg.Logging.Level,
	)
}

func TestNormalizeConfig(t *testing.T) {
	cfg := Config{
		Logging: LoggingConfig{
			Level: "DEBUG",
		},
	}
	Normalize(&cfg)
	assert.Equal(
		t,
		"debug",
		cfg.Logging.Level,
	)
}
