package ingestion

import (
	"context"

	"go.uber.org/fx"

	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/database"
	"github.com/logstorm/api/internal/logger"
)

var Module = fx.Module("ingestion",
	fx.Provide(
		provideRedpandaConfig,
		newPublisher,
		NewIngestionHandler,
		newProcessor,
	),
	fx.Invoke(startProcessor),
)

func provideRedpandaConfig(cfg *config.Config) config.RedpandaConfig {
	return cfg.Redpanda
}

type publisherOut struct {
	fx.Out
	Publisher    Publisher
	DLQPublisher DLQPublisher
}

func newPublisher(lc fx.Lifecycle, cfg config.RedpandaConfig, log *logger.Logger) (publisherOut, error) {
	p, err := NewRedpandaProducer(cfg.Brokers, cfg.Topic, cfg.DLQTopic, *log.Zerolog)
	if err != nil {
		return publisherOut{}, err
	}
	lc.Append(fx.Hook{
		OnStop: func(_ context.Context) error {
			p.Close()
			return nil
		},
	})
	return publisherOut{Publisher: p, DLQPublisher: p}, nil
}

func newProcessor(cfg config.RedpandaConfig, ch database.ClickHouseClient, dlq DLQPublisher, log *logger.Logger) (*LogProcessor, error) {
	return NewLogProcessor(cfg.Brokers, cfg.Topic, cfg.GroupID, ch, dlq, *log.Zerolog)
}

func startProcessor(lc fx.Lifecycle, p *LogProcessor) {
	var cancel context.CancelFunc

	lc.Append(fx.Hook{
		OnStart: func(_ context.Context) error {
			var ctx context.Context
			ctx, cancel = context.WithCancel(context.Background())
			go p.Start(ctx)
			return nil
		},
		OnStop: func(_ context.Context) error {
			cancel()
			return nil
		},
	})
}
