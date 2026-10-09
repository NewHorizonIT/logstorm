CREATE TABLE IF NOT EXISTS logs
(
    id          UUID,
    timestamp   DateTime64(3, 'UTC'),
    project_id  UUID,
    environment LowCardinality(String),
    level       LowCardinality(String),
    service     String             DEFAULT '',
    message     String,
    trace_id    String             DEFAULT '',
    span_id     String             DEFAULT '',
    source      LowCardinality(String) DEFAULT 'sdk',
    attributes  Map(String, String)
)
ENGINE = MergeTree
PARTITION BY toDate(timestamp)
ORDER BY (project_id, timestamp)
SETTINGS index_granularity = 8192;
