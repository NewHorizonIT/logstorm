# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Commands

```bash
# Start all infrastructure (Redpanda, ClickHouse, Postgres, Prometheus, Grafana)
make up

# Build and run the server
make run           # builds to bin/server then runs
go run cmd/server/main.go  # run without building

# Build only
make build         # outputs to bin/server

# Stop / reset infrastructure
make down
make reset         # tears down volumes and restarts fresh

# Load tests (requires k6 installed, server running)
cd tests && ./run_tests.sh smoke    # quick validation
cd tests && ./run_tests.sh load     # comprehensive
cd tests && ./run_tests.sh stress   # find limits
cd tests && ./run_tests.sh spike    # burst traffic
cd tests && ./run_tests.sh soak     # long-running
cd tests && ./run_tests.sh all      # smoke + load + stress + spike
cd tests && ./run_tests.sh generate --count 1000  # generate test log data

# Redpanda topic management (requires infrastructure running)
make topic-create TOPIC_NAME=logs-topic
make topic-list
make topic-consume TOPIC_NAME=logs-topic

# ClickHouse queries
make ch-query QUERY="SELECT count() FROM logstorm.logs"
make ch-file FILE_PATH=deployments/clickhouse/init.sql
```

## Configuration

Config is loaded from a `.env` file (via Viper) with environment variable overrides. Key defaults:

| Variable | Default |
|---|---|
| `SERVER_PORT` | `3123` |
| `KAFKA_BROKERS` | `localhost:9092` |
| `KAFKA_TOPICS` | `logs-topic,dlq-log` |
| `CLICKHOUSE_ADDR` | `localhost:9000` |
| `CLICKHOUSE_DATABASE` | `logstorm` |
| `DB_HOST` / `DB_PORT` | Postgres at `5431` (mapped from container) |

## Architecture

LogStorm is a real-time log analytics pipeline with a single Go binary entrypoint (`cmd/server/main.go`) that runs two concurrent subsystems:

**1. HTTP Ingestion (Gin)**
- `POST /ingestion/` — accepts `domain.Log` JSON, publishes to `logs-topic` on Redpanda; malformed payloads go directly to `dql-topic`
- `GET /metrics` — Prometheus scrape endpoint
- `GET /health` — liveness probe

**2. Background Processor**
- `internal/services/processor` — consumes from `logs-topic`, normalizes logs (uppercase level, lowercase environment, trim whitespace), batches them in-memory (100 logs or 5s flush interval), then bulk-inserts into ClickHouse
- On ClickHouse insert failure: retries with exponential backoff, then sends failed batch to `dlq-log` topic

**Data flow:**
```
HTTP POST /ingestion → Kafka (logs-topic) → Processor (batch/normalize) → ClickHouse
                                                                        ↘ dlq-log (on error)
```

**Infrastructure (docker-compose):**
- **Redpanda** — Kafka-compatible broker on `:9092`; Redpanda Console UI on `:8081`
- **ClickHouse** — OLAP store on `:9000`; initialized from `deployments/clickhouse/init.sql` with `logstorm.logs` table (MergeTree, 90-day TTL), hourly materialized stats, and error-only materialized view
- **PostgreSQL** — on `:5431`; used by the auth service (GORM)
- **Prometheus** — scrapes `:3123/metrics`; on `:9090`
- **Grafana** — on `:3000` (admin/admin); pre-provisioned with Prometheus datasource

**Key packages:**
- `internal/domain` — `Log` struct and `ILogStorm` repository interface
- `internal/infra/kafka` — shared `KafkaClient` (franz-go), producer and consumer wrappers, exponential-backoff retry (`RetryWithBackoff`)
- `internal/infra/clickhouse` — `LogStormRepository`: batch insert with DLQ fallback
- `internal/infra/postgres` — GORM connection for auth
- `internal/services/auth` — in-progress JWT auth service (login/register/refresh/logout routes registered but handlers not yet implemented)
- `internal/observability` — all Prometheus metrics plus Gin middleware that records HTTP request counts and durations
- `pkg` — JSON↔bytes helpers (`JsonToBytes`, `BytesToJson`)

## Auth service (in progress)

`internal/services/auth/` is on the `feat/auth` branch. The router and handler skeleton exist but handler bodies are empty stubs. It is not wired into `main.go` yet.
