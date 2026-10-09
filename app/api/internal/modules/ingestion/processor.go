package ingestion

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/logstorm/api/internal/database"
)

const (
	batchSize     = 100
	batchInterval = 5 * time.Second
)

type DLQPublisher interface {
	PublishDLQ(ctx context.Context, key string, data []byte) error
}

type LogProcessor struct {
	consumer *kgo.Client
	ch       database.ClickHouseClient
	dlq      DLQPublisher
	log      zerolog.Logger
}

func NewLogProcessor(brokers []string, topic, groupID string, ch database.ClickHouseClient, dlq DLQPublisher, log zerolog.Logger) (*LogProcessor, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.ConsumerGroup(groupID),
		kgo.ConsumeTopics(topic),
	)
	if err != nil {
		return nil, fmt.Errorf("create processor consumer: %w", err)
	}

	return &LogProcessor{consumer: client, ch: ch, dlq: dlq, log: log}, nil
}

func (p *LogProcessor) Start(ctx context.Context) {
	p.log.Info().Msg("log processor started")

	for {
		pollCtx, cancel := context.WithTimeout(ctx, batchInterval)
		fetches := p.consumer.PollRecords(pollCtx, batchSize)
		cancel()

		if ctx.Err() != nil {
			p.processFetches(context.Background(), fetches)
			p.consumer.Close()
			p.log.Info().Msg("log processor stopped")
			return
		}

		p.processFetches(ctx, fetches)
	}
}

func (p *LogProcessor) processFetches(ctx context.Context, fetches kgo.Fetches) {
	if fetches.NumRecords() == 0 {
		p.log.Debug().Msg("no log events to process")
		return
	}

	fetches.EachError(func(_ string, _ int32, err error) {
		p.log.Error().Err(err).Msg("consumer fetch error")
	})

	var batch []*LogEvent
	fetches.EachRecord(func(r *kgo.Record) {
		var event LogEvent
		if err := json.Unmarshal(r.Value, &event); err != nil {
			p.log.Error().Err(err).Msg("failed to unmarshal log event")
			if err := p.dlq.PublishDLQ(ctx, string(r.Key), r.Value); err != nil {
				p.log.Error().Err(err).Msg("failed to publish unmarshal failure to DLQ")
			}
			return
		}
		batch = append(batch, &event)
	})

	p.flush(ctx, batch)
}

func (p *LogProcessor) flush(ctx context.Context, batch []*LogEvent) {
	if len(batch) == 0 {
		return
	}

	b, err := p.ch.PrepareBatch(ctx, `INSERT INTO logs
		(id, timestamp, project_id, environment, level, service, message, trace_id, span_id, source, attributes)
		VALUES`)
	if err != nil {
		p.log.Error().Err(err).Msg("prepare clickhouse batch failed")
		p.sendBatchToDLQ(ctx, batch)
		return
	}

	for _, event := range batch {
		if err := b.Append(
			event.ID,
			event.Timestamp,
			event.ProjectID,
			event.Environment,
			event.Level,
			event.Service,
			event.Message,
			event.TraceID,
			event.SpanID,
			event.Source,
			event.Attributes,
		); err != nil {
			p.log.Error().Err(err).Str("event_id", event.ID.String()).Msg("append to clickhouse batch failed")
		}
	}

	if err := b.Send(); err != nil {
		p.log.Error().Err(err).Int("count", len(batch)).Msg("clickhouse batch send failed")
		p.sendBatchToDLQ(ctx, batch)
		return
	}

	p.log.Debug().Int("count", len(batch)).Msg("flushed log batch to clickhouse")
}

func (p *LogProcessor) sendBatchToDLQ(ctx context.Context, batch []*LogEvent) {
	for _, event := range batch {
		data, err := json.Marshal(event)
		if err != nil {
			p.log.Error().Err(err).Str("event_id", event.ID.String()).Msg("failed to marshal event for DLQ")
			continue
		}
		if err := p.dlq.PublishDLQ(ctx, event.ID.String(), data); err != nil {
			p.log.Error().Err(err).Str("event_id", event.ID.String()).Msg("failed to publish event to DLQ")
		}
	}
}
