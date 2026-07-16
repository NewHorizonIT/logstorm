package logger

import (
	"os"
	"path/filepath"

	"github.com/logstorm/api/internal/config"
	"gopkg.in/natefinch/lumberjack.v2"
)

func NewRotatingWriter(
	config config.LoggingConfig,
) (*lumberjack.Logger, error) {
	if err := os.MkdirAll(
		filepath.Dir(config.FilePath),
		0755,
	); err != nil {
		return nil, err
	}

	return &lumberjack.Logger{
		Filename:   config.FilePath,
		MaxSize:    config.RotationMaxSize,
		MaxBackups: config.RotationMaxBackups,
		MaxAge:     config.RotationMaxAge,
		Compress:   config.CompressionEnabled,
	}, nil
}
