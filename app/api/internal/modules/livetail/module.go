package livetail

import (
	"context"

	"go.uber.org/fx"

	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/logger"
)

var Module = fx.Module("livetail",
	fx.Provide(
		NewHub,
		newConsumer,
		NewHandler,
	),
	fx.Invoke(startConsumer),
)

func newConsumer(cfg *config.Config, hub *Hub, log *logger.Logger) (*LiveTailConsumer, error) {
	return NewLiveTailConsumer(cfg.Redpanda, hub, *log.Zerolog)
}

func startConsumer(lc fx.Lifecycle, c *LiveTailConsumer) {
	var cancel context.CancelFunc
	done := make(chan struct{})

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go func() {
				defer close(done)
				c.Start(ctx)
			}()
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			<-done
			return nil
		},
	})
}
