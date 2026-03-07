package kafka

import (
	"context"

	"github.com/twmb/franz-go/pkg/kgo"
)

type KafkaProducer struct {
	client *kgo.Client
}

func NewKafkaProducer(client *kgo.Client) *KafkaProducer {
	return &KafkaProducer{
		client: client,
	}
}

func (kp *KafkaProducer) Publish(ctx context.Context, topic string, key []byte, message []byte) error {
	record := &kgo.Record{
		Topic: topic,
		Key:   key,
		Value: message,
	}

	return kp.client.ProduceSync(ctx, record).FirstErr()
}
