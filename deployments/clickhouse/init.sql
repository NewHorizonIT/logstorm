-- LogStorm ClickHouse Schema
-- Database và tables cho log analytics system

-- Tạo database
CREATE DATABASE IF NOT EXISTS logstorm;

-- Sử dụng database
USE logstorm;

-- Bảng logs chính - lưu trữ tất cả logs từ các services
CREATE TABLE IF NOT EXISTS logs
(
    id          UInt64,
    message     String,
    trace_id    String,
    environment LowCardinality(String),  -- dev, staging, production
    level       LowCardinality(String),  -- debug, info, warn, error, fatal
    service     LowCardinality(String),  -- tên service
    timestamp   DateTime64(3),           -- timestamp với milliseconds precision

-- Metadata cho analytics
inserted_at DateTime DEFAULT now()
)
ENGINE = MergeTree()
PARTITION BY toYYYYMM(timestamp)         -- Partition theo tháng
ORDER BY (service, level, timestamp, id)  -- Tối ưu cho queries thường gặp
TTL timestamp + INTERVAL 90 DAY           -- Tự động xóa logs sau 90 ngày
SETTINGS index_granularity = 8192;

-- Index phụ để tăng tốc queries
ALTER TABLE logs
ADD INDEX idx_trace_id trace_id TYPE bloom_filter GRANULARITY 4;

ALTER TABLE logs
ADD INDEX idx_message message TYPE tokenbf_v1 (32768, 3, 0) GRANULARITY 4;

-- Materialized View cho thống kê theo service và level (mỗi giờ)
CREATE TABLE IF NOT EXISTS logs_hourly_stats (
    hour DateTime,
    service LowCardinality (String),
    level LowCardinality (String),
    environment LowCardinality (String),
    log_count UInt64
) ENGINE = SummingMergeTree ()
PARTITION BY
    toYYYYMM (hour)
ORDER BY (
        hour, service, level, environment
    );

CREATE MATERIALIZED VIEW IF NOT EXISTS logs_hourly_stats_mv TO logs_hourly_stats AS
SELECT
    toStartOfHour (timestamp) AS hour,
    service,
    level,
    environment,
    count() AS log_count
FROM logs
GROUP BY
    hour,
    service,
    level,
    environment;

-- Materialized View cho error tracking (chỉ error và fatal)
CREATE TABLE IF NOT EXISTS error_logs (
    id UInt64,
    message String,
    trace_id String,
    environment LowCardinality (String),
    level LowCardinality (String),
    service LowCardinality (String),
    timestamp DateTime64 (3)
) ENGINE = MergeTree ()
PARTITION BY
    toYYYYMMDD (timestamp)
ORDER BY (service, timestamp, id) TTL timestamp + INTERVAL 30 DAY;

CREATE MATERIALIZED VIEW IF NOT EXISTS error_logs_mv TO error_logs AS
SELECT
    id,
    message,
    trace_id,
    environment,
    level,
    service,
    timestamp
FROM logs
WHERE
    level IN (
        'error',
        'fatal',
        'ERROR',
        'FATAL'
    );

-- View để query logs theo trace_id (distributed tracing)
CREATE VIEW IF NOT EXISTS logs_by_trace AS
SELECT *
FROM logs
ORDER BY timestamp ASC;