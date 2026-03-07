package kafka

import (
	"context"
	"log/slog"

	"github.com/NewHorizonIT/logstorm/internal/services/event"
	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaConsumer struct {
	client *kgo.Client
}

func NewKafkaConsumer(client *kgo.Client) *KafkaConsumer {
	return &KafkaConsumer{
		client: client,
	}
}

// Consume starts consuming messages from topic and calls handler for each message
func (kc *KafkaConsumer) Consume(ctx context.Context, topic string, handler event.MessageHandler) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			fetches := kc.client.PollFetches(ctx)

			if errs := fetches.Errors(); len(errs) > 0 {
				for _, err := range errs {
					slog.Error("fetch error", "topic", err.Topic, "partition", err.Partition, "error", err.Err)
				}
				continue
			}

			fetches.EachRecord(func(r *kgo.Record) {
				if err := handler(ctx, r.Key, r.Value); err != nil {
					slog.Error("handler error", "topic", r.Topic, "partition", r.Partition, "offset", r.Offset, "error", err)
				}
			})
		}
	}
}

func (kc *KafkaConsumer) Stop() error {
	kc.client.Close()
	return nil
}
