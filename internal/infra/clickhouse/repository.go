package clickhouse

import (
	"context"
	"encoding/json"
	"log/slog"
	"time"

	"github.com/ClickHouse/clickhouse-go/v2"
	"github.com/NewHorizonIT/logstorm/internal/domain"
	"github.com/NewHorizonIT/logstorm/internal/infra/kafka"
	"github.com/NewHorizonIT/logstorm/internal/observability"
	"github.com/NewHorizonIT/logstorm/internal/services/event"
)

const DLQTopic = "dlq-log"

type LogStormRepository struct {
	conn      clickhouse.Conn
	cnf       kafka.RetryCnf
	publisher event.IPublisher
}

func NewLogStormRepository(conn clickhouse.Conn, cnf kafka.RetryCnf, publisher event.IPublisher) *LogStormRepository {
	return &LogStormRepository{
		conn:      conn,
		cnf:       cnf,
		publisher: publisher,
	}
}

func (r *LogStormRepository) InsertLogs(ctx context.Context, logs []domain.Log) error {
	start := time.Now()
	observability.ClickHouseInsertBatchSize.Observe(float64(len(logs)))

	batch, err := r.conn.PrepareBatch(ctx, `INSERT INTO logstorm.logs (id, message, trace_id, environment, level, service, timestamp)`)
	if err != nil {
		observability.ClickHouseInsertTotal.WithLabelValues("error").Inc()
		r.sendToDLQ(ctx, logs, err)
		return err
	}

	for _, log := range logs {
		if err := batch.Append(
			uint64(log.ID),
			log.Message,
			log.TraceID,
			log.Environment,
			log.Level,
			log.Service,
			log.Timestamp,
		); err != nil {
			observability.ClickHouseInsertTotal.WithLabelValues("error").Inc()
			r.sendToDLQ(ctx, logs, err)
			return err
		}
	}

	// Retry on Send (network operation) not Append (local operation)
	err = kafka.RetryWithBackoff(ctx, r.cnf, func() error {
		return batch.Send()
	})

	if err != nil {
		observability.ClickHouseInsertTotal.WithLabelValues("error").Inc()
		r.sendToDLQ(ctx, logs, err)
		return err
	}

	observability.ClickHouseInsertTotal.WithLabelValues("success").Inc()
	observability.ClickHouseInsertDuration.Observe(time.Since(start).Seconds())

	return nil
}

// sendToDLQ pushes failed logs to dead letter queue for later processing
func (r *LogStormRepository) sendToDLQ(ctx context.Context, logs []domain.Log, originalErr error) {
	observability.DLQMessagesTotal.WithLabelValues("clickhouse_error").Inc()

	dlqPayload := struct {
		Logs  []domain.Log `json:"logs"`
		Error string       `json:"error"`
	}{
		Logs:  logs,
		Error: originalErr.Error(),
	}

	data, err := json.Marshal(dlqPayload)
	if err != nil {
		slog.Error("failed to marshal DLQ payload", "error", err)
		return
	}

	if err := r.publisher.Publish(ctx, DLQTopic, nil, data); err != nil {
		slog.Error("failed to send to DLQ", "error", err, "original_error", originalErr)
	}
}
