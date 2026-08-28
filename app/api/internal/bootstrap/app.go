package bootstrap

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"

	"github.com/jackc/pgx/v5/pgxpool"
	"go.uber.org/fx"

	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/database"
	"github.com/logstorm/api/internal/logger"
)

var Module = fx.Module("bootstrap",
	fx.Supply(config.LoaderOptions{ConfigFile: "configs/config.yaml"}),
	fx.Provide(
		config.Load,
		provideLoggingConfig,
		provideDatabaseConfig,
		provideClickHouseConfig,
		provideLogger,
		providePostgres,
		provideClickHouse,
		providePgPool,
		SetupRouter,
	),
	fx.Invoke(startServer),
)

// -- Config extractors -------------------------------------------------------

func provideLoggingConfig(cfg *config.Config) config.LoggingConfig {
	return cfg.Logging
}

func provideDatabaseConfig(cfg *config.Config) config.DatabaseConfig {
	return cfg.Database
}

func provideClickHouseConfig(cfg *config.Config) config.ClickHouseConfig {
	return cfg.ClickHouse
}

// -- Infrastructure providers with lifecycle ---------------------------------

func provideLogger(lc fx.Lifecycle, cfg config.LoggingConfig) (*logger.Logger, error) {
	l, err := logger.New(cfg)
	if err != nil {
		return nil, err
	}
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return l.Close()
		},
	})
	return l, nil
}

func providePostgres(lc fx.Lifecycle, cfg config.DatabaseConfig) (*database.Postgres, error) {
	pg, err := database.Connect(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect postgres: %w", err)
	}
	if err := pg.Ping(); err != nil {
		return nil, fmt.Errorf("ping postgres: %w", err)
	}
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			pg.Close()
			return nil
		},
	})
	return pg, nil
}

func provideClickHouse(lc fx.Lifecycle, cfg config.ClickHouseConfig) (database.ClickHouseClient, error) {
	ch, err := database.NewClickHouse(cfg)
	if err != nil {
		return nil, fmt.Errorf("connect clickhouse: %w", err)
	}
	if err := ch.Ping(); err != nil {
		ch.Close()
		return nil, fmt.Errorf("ping clickhouse: %w", err)
	}
	if err := ch.HealthCheck(); err != nil {
		ch.Close()
		return nil, fmt.Errorf("clickhouse health check: %w", err)
	}
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			return ch.Close()
		},
	})
	return ch, nil
}

func providePgPool(pg *database.Postgres) *pgxpool.Pool {
	return pg.Pool
}

// -- HTTP server lifecycle ----------------------------------------------------

func startServer(lc fx.Lifecycle, cfg *config.Config, log *logger.Logger, router http.Handler) {
	addr := net.JoinHostPort(cfg.Server.Host, strconv.Itoa(cfg.Server.Port))
	srv := &http.Server{Addr: addr, Handler: router}

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			log.Zerolog.Info().Str("address", addr).Msg("server starting")
			go func() {
				if err := srv.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
					log.Zerolog.Error().Err(err).Msg("server error")
				}
			}()
			return nil
		},
		OnStop: func(ctx context.Context) error {
			log.Zerolog.Info().Msg("server stopping")
			return srv.Shutdown(ctx)
		},
	})
}
