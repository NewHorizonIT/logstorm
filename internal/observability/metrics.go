package observability

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "path", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "logstorm_http_request_duration_seconds",
			Help:    "HTTP request duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
		[]string{"method", "path"},
	)

	// Ingestion Metrics
	LogsIngestedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_logs_ingested_total",
			Help: "Total number of logs ingested",
		},
		[]string{"service", "level"},
	)

	IngestionBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "logstorm_ingestion_batch_size",
			Help:    "Size of ingestion batches",
			Buckets: []float64{1, 10, 50, 100, 500, 1000, 5000},
		},
	)

	// Processor Metrics
	LogsProcessedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_logs_processed_total",
			Help: "Total number of logs processed",
		},
		[]string{"service", "level", "environment"},
	)

	ProcessorBufferSize = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "logstorm_processor_buffer_size",
			Help: "Current size of processor buffer",
		},
	)

	ProcessorFlushTotal = promauto.NewCounter(
		prometheus.CounterOpts{
			Name: "logstorm_processor_flush_total",
			Help: "Total number of flush operations",
		},
	)

	ProcessorFlushDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "logstorm_processor_flush_duration_seconds",
			Help:    "Duration of flush operations in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	// ClickHouse Metrics
	ClickHouseInsertTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_clickhouse_insert_total",
			Help: "Total number of ClickHouse insert operations",
		},
		[]string{"status"}, // success, error
	)

	ClickHouseInsertDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "logstorm_clickhouse_insert_duration_seconds",
			Help:    "ClickHouse insert duration in seconds",
			Buckets: prometheus.DefBuckets,
		},
	)

	ClickHouseInsertBatchSize = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "logstorm_clickhouse_insert_batch_size",
			Help:    "Size of ClickHouse insert batches",
			Buckets: []float64{1, 10, 50, 100, 500, 1000},
		},
	)

	// Kafka Metrics
	KafkaProduceTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_kafka_produce_total",
			Help: "Total number of Kafka produce operations",
		},
		[]string{"topic", "status"},
	)

	KafkaConsumeTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_kafka_consume_total",
			Help: "Total number of Kafka consume operations",
		},
		[]string{"topic", "status"},
	)

	// DLQ Metrics
	DLQMessagesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_dlq_messages_total",
			Help: "Total number of messages sent to DLQ",
		},
		[]string{"reason"},
	)

	// Retry Metrics
	RetryAttemptsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "logstorm_retry_attempts_total",
			Help: "Total number of retry attempts",
		},
		[]string{"operation", "attempt"},
	)
)
