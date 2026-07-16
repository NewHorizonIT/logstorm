package logger

import (
	"context"

	"github.com/rs/zerolog"
)

type contextKey string

const loggerContextKey contextKey = "logger"

func WithLogger(ctx context.Context, log zerolog.Logger) context.Context {
	return context.WithValue(ctx, loggerContextKey, log)
}

func FromContext(ctx context.Context) zerolog.Logger {
	log, ok := ctx.Value(loggerContextKey).(zerolog.Logger)

	if !ok {
		return zerolog.Nop()
	}
	return log
}
