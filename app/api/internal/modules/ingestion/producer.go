package ingestion

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"github.com/rs/zerolog"
	"github.com/twmb/franz-go/pkg/kadm"
	"github.com/twmb/franz-go/pkg/kerr"
	"github.com/twmb/franz-go/pkg/kgo"
)

type RedpandaProducer struct {
	client   *kgo.Client
	topic    string
	dlqTopic string
	log      zerolog.Logger
}

func NewRedpandaProducer(brokers []string, topic, dlqTopic string, log zerolog.Logger) (*RedpandaProducer, error) {
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

	if err := ensureTopic(client, dlqTopic); err != nil {
		client.Close()
		return nil, fmt.Errorf("ensure dlq topic %q: %w", dlqTopic, err)
	}

	return &RedpandaProducer{client: client, topic: topic, dlqTopic: dlqTopic, log: log}, nil
}

func (p *RedpandaProducer) Publish(ctx context.Context, event *LogEvent) error {
	payload, err := json.Marshal(event)
	if err != nil {
		return fmt.Errorf("marshal log event: %w", err)
	}

	record := &kgo.Record{
		Topic: p.topic,
		Key:   []byte(event.ProjectID.String()),
		Value: payload,
	}

	// Use context.Background() — request context may be canceled before
	// the async callback fires (handler returns 202 before produce completes).
	p.client.Produce(context.Background(), record, func(r *kgo.Record, err error) {
		if err != nil {
			p.log.Error().Err(err).
				Str("topic", r.Topic).
				Str("project_id", string(r.Key)).
				Msg("failed to publish log event to redpanda")
		}
	})

	return nil
}

func (p *RedpandaProducer) Close() {
	if err := p.client.Flush(context.Background()); err != nil {
		p.log.Error().Err(err).Msg("redpanda producer flush on close failed")
	}
	p.client.Close()
}

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

func (p *RedpandaProducer) PublishDLQ(_ context.Context, key string, data []byte) error {
	record := &kgo.Record{
		Topic: p.dlqTopic,
		Key:   []byte(key),
		Value: data,
	}

	p.client.Produce(context.Background(), record, func(r *kgo.Record, err error) {
		if err != nil {
			p.log.Error().Err(err).
				Str("topic", r.Topic).
				Msg("failed to publish to DLQ")
		}
	})

	return nil
}
