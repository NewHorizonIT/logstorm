package logger

import (
	"io"
	"os"
	"time"

	"github.com/logstorm/api/internal/config"
	"github.com/rs/zerolog"
)

type Logger struct {
	Zerolog    *zerolog.Logger
	fileWriter io.Closer
}

func New(config config.LoggingConfig) (*Logger, error) {
	level, err := zerolog.ParseLevel(config.Level)

	if err != nil {
		level = zerolog.InfoLevel
	}

	zerolog.SetGlobalLevel(level)
	writers := make([]io.Writer, 0, 2)
	var fileWriter io.Closer

	if config.Environment == "development" {
		consoleWriter := zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.RFC3339,
		}
		writers = append(
			writers,
			consoleWriter,
		)
	} else {
		writers = append(
			writers,
			os.Stdout,
		)
	}

	if config.FileEnabled {
		rotatingWriter, err := NewRotatingWriter(
			config,
		)
		if err != nil {
			return nil, err
		}
		writers = append(
			writers,
			rotatingWriter,
		)
		fileWriter = rotatingWriter
	}

	writer := zerolog.MultiLevelWriter(
		writers...,
	)
	log := zerolog.New(writer).
		With().
		Timestamp().
		Str(FieldEnvironment, config.Environment).
		Logger()
	return &Logger{
		Zerolog:    &log,
		fileWriter: fileWriter,
	}, nil
}

func (l *Logger) Close() error {
	if l.fileWriter == nil {
		return nil
	}
	return l.fileWriter.Close()
}
