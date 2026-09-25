package messaging

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"

	"github.com/logstorm/api/internal/modules/ingestion"
)

type RedpandaProducer struct {
	client *kgo.Client
	topic  string
	log    zerolog.Logger
}

func NewRedpandaProducer(brokers []string, topic string, log zerolog.Logger) (*RedpandaProducer, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(brokers...),
		kgo.RequiredAcks(kgo.LeaderAck()),
	)
	if err != nil {
		return nil, fmt.Errorf("create redpanda client: %w", err)
	}

	if err := ensureTopic(client, topic); err != nil {
		client.Close()
		return nil, fmt.Errorf("ensure topic %q: %w", topic, err)
	}

	return &RedpandaProducer{client: client, topic: topic, log: log}, nil
}

func (p *RedpandaProducer) Publish(ctx context.Context, event *ingestion.LogEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal log event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.ProjectID.String()),
		Value: payload,
	}

	p.client.Produce(ctx, record, func(r *kgo.Record, err error) {
		if err != nil {
			p.log.Error().Err(err).
				Str("topic", r.Topic).
				Str("project_id", string(r.Key)).
				Msg("failed to publish log event to redpanda")
		}
	})

	return nil
}

// Close flushes pending records and shuts down the client.
func (p *RedpandaProducer) Close() {
	if err := p.client.Flush(context.Background()); err != nil {
		p.log.Error().Err(err).Msg("redpanda producer flush on close failed")
	}
	p.client.Close()
}

// ensureTopic creates the topic if it does not already exist.
func ensureTopic(client *kgo.Client, topic string) error {
	adm := kadm.NewClient(client)

	resp, err := adm.CreateTopics(context.Background(), 3, 1, nil, topic)
	if err != nil {
		return err
	}

	for _, t := range resp {
		if t.Err != nil && !errors.Is(t.Err, kerr.TopicAlreadyExists) {
			return fmt.Errorf("create topic: %w", t.Err)
		}
	}

	return nil
}
