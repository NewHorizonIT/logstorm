package producer

import (
	"context"
	"sync"
	"time"

	"github.com/NewHorizonIT/logstorm-ingestion/internal/encoder"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/model"
	"github.com/NewHorizonIT/logstorm-ingestion/internal/observability"
	"github.com/segmentio/kafka-go"
)

type KafkaProducer struct {
	writer     *kafka.Writer
	queue      chan model.LogEvent
	wg         sync.WaitGroup
	flushEvery time.Duration
	wokerCount int
	batchSize  int
	encoder    encoder.Encoder
}

func NewKafkaProducer(broker, topic string, bufferSize int, workerCount int) *KafkaProducer {
	writer := &kafka.Writer{
		Addr:         kafka.TCP(broker),
		Topic:        topic,
		RequiredAcks: kafka.RequireAll,
		BatchTimeout: 10 * time.Millisecond,
		BatchSize:    1000,
		Compression:  kafka.Snappy,
	}

	return &KafkaProducer{
		writer:     writer,
		queue:      make(chan model.LogEvent, bufferSize),
		batchSize:  500,
		wokerCount: workerCount,
		flushEvery: 10 * time.Millisecond,
		encoder:    encoder.JSONEncoder{},
	}
}

func (k *KafkaProducer) Enqueue(event model.LogEvent) error {
	select {
	case k.queue <- event:
		observability.QueueSize.Set(float64(len(k.queue)))
		return nil
	default:
		return context.DeadlineExceeded
	}
}

func (k *KafkaProducer) Start(ctx context.Context) {
	for i := 0; i < k.wokerCount; i++ {
		k.wg.Add(1)
		go k.worker(ctx)
	}
}

func (k *KafkaProducer) sendWithRetry(ctx context.Context, messages []kafka.Message) {
	backoff := 100 * time.Millisecond

	// Setup 5 retries with "exponential backoff"
	retryCount := 5
	for i := 0; i < retryCount; i++ {
		err := k.writer.WriteMessages(ctx, messages...)
		if err == nil {
			return
		}

		time.Sleep(backoff)
		backoff *= 2
	}
}

func (k *KafkaProducer) Close() {
	close(k.queue)
	k.wg.Wait()
	k.writer.Close()
}
