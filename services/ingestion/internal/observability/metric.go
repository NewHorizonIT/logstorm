package observability

import "github.com/prometheus/client_golang/prometheus"

var (
	HTTPRequestsTotal = prometheus.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total HTTP requests",
		},
		[]string{"method", "status"},
	)

	HTTPRequestDuration = prometheus.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "HTTP request latency",
			Buckets: prometheus.DefBuckets,
		},
	)

	QueueSize = prometheus.NewGauge(
		prometheus.GaugeOpts{
			Name: "ingestion_queue_size",
			Help: "Current queue size",
		},
	)

	KafkaProduceTotal = prometheus.NewCounter(
		prometheus.CounterOpts{
			Name: "kafka_produce_total",
			Help: "Total messages sent to Kafka",
		},
	)
)

func Init() {
	prometheus.MustRegister(
		HTTPRequestsTotal,
		HTTPRequestDuration,
		QueueSize,
		KafkaProduceTotal,
	)
}
