package processor

import (
	"context"
	"log/slog"
	"strings"
	"sync"
	"time"

	"github.com/NewHorizonIT/logstorm/internal/domain"
	"github.com/NewHorizonIT/logstorm/internal/observability"
	"github.com/NewHorizonIT/logstorm/internal/services/event"
	"github.com/NewHorizonIT/logstorm/pkg"
)

const (
	LogsTopic     = "logs-topic"
	DLQTopic      = "dlq-log"
	BatchSize     = 100             // Number of logs per batch
	FlushInterval = 5 * time.Second // Max time to wait before flushing
)

type Processor struct {
	consumer   event.IConsumer
	repository domain.ILogStorm
	publisher  event.IPublisher

	// Batch buffer
	mu     sync.Mutex
	buffer []domain.Log
}

func NewProcessor(consumer event.IConsumer, repository domain.ILogStorm, publisher event.IPublisher) *Processor {
	return &Processor{
		consumer:   consumer,
		repository: repository,
		publisher:  publisher,
		buffer:     make([]domain.Log, 0, BatchSize),
	}
}

// Start begins consuming and processing logs
func (p *Processor) Start(ctx context.Context) error {
	// Start flush ticker for time-based flushing
	go p.flushTicker(ctx)

	// Start consuming
	return p.consumer.Consume(ctx, LogsTopic, p.handleMessage)
}

func (p *Processor) handleMessage(ctx context.Context, key, value []byte) error {
	var rawLog domain.Log
	if err := pkg.BytesToJson(value, &rawLog); err != nil {
		slog.Error("failed to parse log", "error", err)
		observability.KafkaConsumeTotal.WithLabelValues(LogsTopic, "error").Inc()
		if pubErr := p.publisher.Publish(ctx, DLQTopic, key, value); pubErr != nil {
			slog.Error("failed to send unparseable message to DLQ", "error", pubErr)
		}
		return err
	}

	observability.KafkaConsumeTotal.WithLabelValues(LogsTopic, "success").Inc()

	// Process the log
	processedLog := p.processLog(rawLog)

	// Record processed metrics
	observability.LogsProcessedTotal.WithLabelValues(
		processedLog.Service,
		processedLog.Level,
		processedLog.Environment,
	).Inc()

	// Add to buffer
	p.addToBuffer(ctx, processedLog)

	return nil
}

// Future enhancements: using spark for distributed processing, implementing more complex transformations, adding metrics and monitoring
// processLog applies transformations to the log
func (p *Processor) processLog(log domain.Log) domain.Log {
	// Normalize level to uppercase
	log.Level = strings.ToUpper(log.Level)

	// Trim whitespace from message
	log.Message = strings.TrimSpace(log.Message)

	// Set timestamp if not provided
	if log.Timestamp == 0 {
		log.Timestamp = time.Now().UnixMilli()
	}

	// Normalize environment
	log.Environment = strings.ToLower(log.Environment)

	// Trim service name
	log.Service = strings.TrimSpace(log.Service)

	return log
}

// addToBuffer adds log to buffer and flushes if batch is full
func (p *Processor) addToBuffer(ctx context.Context, log domain.Log) {
	p.mu.Lock()
	p.buffer = append(p.buffer, log)

	// Update buffer size metric
	observability.ProcessorBufferSize.Set(float64(len(p.buffer)))

	if len(p.buffer) >= BatchSize {
		logs := p.buffer
		p.buffer = make([]domain.Log, 0, BatchSize)
		observability.ProcessorBufferSize.Set(0)
		p.mu.Unlock()

		p.flush(ctx, logs)
		return
	}
	p.mu.Unlock()
}

// flushTicker periodically flushes the buffer
func (p *Processor) flushTicker(ctx context.Context) {
	ticker := time.NewTicker(FlushInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			// Final flush on shutdown
			p.mu.Lock()
			logs := p.buffer
			p.buffer = nil
			p.mu.Unlock()

			if len(logs) > 0 {
				p.flush(context.Background(), logs)
			}
			return

		case <-ticker.C:
			p.mu.Lock()
			if len(p.buffer) > 0 {
				logs := p.buffer
				p.buffer = make([]domain.Log, 0, BatchSize)
				p.mu.Unlock()

				p.flush(ctx, logs)
			} else {
				p.mu.Unlock()
			}
		}
	}
}

// flush inserts logs into ClickHouse
func (p *Processor) flush(ctx context.Context, logs []domain.Log) {
	if len(logs) == 0 {
		return
	}

	start := time.Now()
	slog.Info("flushing logs to ClickHouse", "count", len(logs))

	observability.ProcessorFlushTotal.Inc()

	if err := p.repository.InsertLogs(ctx, logs); err != nil {
		slog.Error("failed to insert logs", "error", err, "count", len(logs))
	}

	observability.ProcessorFlushDuration.Observe(time.Since(start).Seconds())
}
