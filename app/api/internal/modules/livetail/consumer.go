package livetail

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/logstorm/api/internal/config"
	"github.com/logstorm/api/internal/modules/ingestion"
)

const StalenessThreshold = 60 * time.Second

type Broadcaster interface {
	Broadcast(events []*ingestion.LogEvent)
}

type LiveTailConsumer struct {
	consumer    *kgo.Client
	broadcaster Broadcaster
	log         zerolog.Logger
}

func NewLiveTailConsumer(cfg config.RedpandaConfig, hub *Hub, log zerolog.Logger) (*LiveTailConsumer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(cfg.Brokers...),
		kgo.ConsumerGroup(cfg.LiveTailGroupID),
		kgo.ConsumeTopics(cfg.Topic),
		kgo.ConsumeResetOffset(kgo.NewOffset().AtEnd()),
	)
	if err != nil {
		return nil, fmt.Errorf("create livetail consumer: %w", err)
	}

	return &LiveTailConsumer{
		consumer:    client,
		broadcaster: hub,
		log:         log,
	}, nil
}

func (c *LiveTailConsumer) Start(ctx context.Context) {
	c.log.Info().Msg("live tail consumer started")

	for {
		fetches := c.consumer.PollFetches(ctx)

		if ctx.Err() != nil {
			c.consumer.Close()
			c.log.Info().Msg("live tail consumer stopped")
			return
		}

		fetches.EachError(func(_ string, _ int32, err error) {
			c.log.Error().Err(err).Msg("live tail: fetch error")
		})

		var records []*kgo.Record
		fetches.EachRecord(func(r *kgo.Record) {
			records = append(records, r)
		})

		c.processRecords(records)
	}
}

func (c *LiveTailConsumer) processRecords(records []*kgo.Record) {
	if len(records) == 0 {
		return
	}

	now := time.Now().UTC()
	var batch []*ingestion.LogEvent

	for _, r := range records {
		var event ingestion.LogEvent
		if err := json.Unmarshal(r.Value, &event); err != nil {
			c.log.Warn().Err(err).Msg("live tail: failed to unmarshal log event, skipping")
			continue
		}

		if now.Sub(event.Timestamp) > StalenessThreshold {
			c.log.Debug().
				Str("event_id", event.ID.String()).
				Time("event_timestamp", event.Timestamp).
				Dur("age", now.Sub(event.Timestamp)).
				Msg("live tail: dropping stale event")
			continue
		}

		batch = append(batch, &event)
	}

	if len(batch) > 0 {
		c.broadcaster.Broadcast(batch)
	}
}
