package kafka

import (
	"time"

	"github.com/twmb/franz-go/pkg/kgo"
)

type Config struct {
	Brokers       []string // List of Kafka broker addresses
	ClientID      string   // Unique identifier for the client
	ConsumerGroup string   // Only needed for consumers
	Topics        []string // List of topics to consume from or produce to

	Linger       time.Duration // Time to wait before sending a batch of messages
	BatchSize    int32         // Maximum size of a batch of messages
	Compression  string        // Compression codec (e.g., "gzip", "snappy", "lz4")
	RetryTimeout time.Duration // Time to wait before retrying a failed message send
}

// Example usage:
/*
config := Config{
	Brokers:       []string{"localhost:9092"},
	ClientID:      "my-app",
	ConsumerGroup: "my-group",
	Topics:        []string{"my-topic"},
	Linger:        100 * time.Millisecond,
	BatchSize:     100,
	Compression:   "gzip",
	RetryTimeout:  5 * time.Second,
}
*/

func InitKafkaClient(config Config) (*kgo.Client, error) {
	client, err := kgo.NewClient(
		kgo.SeedBrokers(config.Brokers...),
		kgo.ClientID(config.ClientID),
		kgo.RequiredAcks(kgo.AllISRAcks()),

		// Consumer
		kgo.ProducerLinger(config.Linger),
		kgo.ProducerBatchMaxBytes(config.BatchSize),
		kgo.RetryTimeout(config.RetryTimeout),
		kgo.ProducerBatchCompression(kgo.Lz4Compression()),

		// Conusmer
		kgo.ConsumerGroup(config.ConsumerGroup),
		kgo.DisableAutoCommit(),
		kgo.ConsumeTopics(config.Topics...),
	)
	if err != nil {
		return nil, err
	}

	return client, nil
}
