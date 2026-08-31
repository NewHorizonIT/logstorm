package auth

import (
	"context"
	"time"

	"go.uber.org/fx"

	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
)

var Module = fx.Module("auth",
	fx.Provide(
		provideAuthConfig,
		NewTokenService,
		NewAuthService,
	),
	fx.Invoke(startCleanupGoroutine),
)

func provideAuthConfig(cfg *config.Config) config.AuthConfig {
	return cfg.Auth
}

func startCleanupGoroutine(lc fx.Lifecycle, repo AuthRepository, log *logger.Logger) {
	var cancel context.CancelFunc

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go func() {
				ticker := time.NewTicker(6 * time.Hour)
				defer ticker.Stop()
				for {
					select {
					case <-ticker.C:
						if err := repo.DeleteExpiredTokens(context.Background()); err != nil {
							log.Zerolog.Error().Err(err).Msg("auth: failed to delete expired refresh tokens")
						}
					case <-ctx.Done():
						return
					}
				}
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}
