# LogStorm

Real-time log analytics platform cho microservices. Xây dựng với Go, Kafka (Redpanda), ClickHouse.

## Features

- **High-throughput ingestion** - Nhận log từ nhiều service qua REST API
- **Event-driven processing** - Kafka/Redpanda làm message broker với retry & DLQ
- **Fast analytics** - ClickHouse cho query nhanh trên dữ liệu lớn
- **Observability** - Prometheus metrics + Grafana dashboards

## Tech Stack

| Component      | Technology                  |
| -------------- | --------------------------- |
| Language       | Go 1.25                     |
| Web Framework  | Gin                         |
| Message Broker | Redpanda (Kafka-compatible) |
| Database       | ClickHouse                  |
| Monitoring     | Prometheus + Grafana        |

## Quick Start

```bash
# Start infrastructure
cd deployments && docker compose up -d

# Create Kafka topics
make topic-create TOPIC_NAME=logs
make topic-create TOPIC_NAME=logs-dlq

# Run server
go run cmd/server/main.go
```

## API

```bash
# Ingest logs
POST /api/v1/logs
Content-Type: application/json

{"level": "info", "message": "User logged in", "service": "auth"}
```

## Project Structure

```
cmd/server/       # Application entrypoint
internal/
  domain/         # Domain models
  infra/          # ClickHouse, Kafka clients
  services/       # Ingestion, processor, event handlers
  observability/  # Metrics, middleware
deployments/      # Docker Compose, Prometheus, Grafana configs
tests/            # k6 load tests
```
