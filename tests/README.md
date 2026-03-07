# LogStorm Performance Tests

Performance và load tests cho LogStorm sử dụng [k6](https://k6.io/).

## Prerequisites

```bash
# Install k6
# macOS
brew install k6

# Linux (Debian/Ubuntu)
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Windows
choco install k6
```

## Test Types

| Test             | Mục đích                      | Duration | VUs     |
| ---------------- | ----------------------------- | -------- | ------- |
| `smoke_test.js`  | Kiểm tra hệ thống hoạt động   | 1m       | 10      |
| `load_test.js`   | Full test với nhiều scenarios | 20m      | 50-500  |
| `stress_test.js` | Tìm giới hạn hệ thống         | 30m      | 100-500 |
| `spike_test.js`  | Test burst traffic            | 3m       | 10-1000 |
| `soak_test.js`   | Detect memory leaks           | 4h       | 50      |

## Quick Start

```bash
cd tests

# 1. Start LogStorm server first
go run ../cmd/server/main.go

# 2. Run smoke test
./run_tests.sh smoke

# 3. Run full load test
./run_tests.sh load
```

## Manual Commands

```bash
# Basic run
k6 run smoke_test.js

# Custom URL
k6 run -e BASE_URL=http://prod:3123 load_test.js

# Output to JSON
k6 run --out json=results.json stress_test.js

# Run specific scenario only
k6 run --scenario smoke load_test.js
```

## Log Generator

Generate logs để test thủ công:

```bash
# Generate 100 random logs
node generate_logs.js

# Generate 1000 logs
node generate_logs.js --count 1000

# Generate error-heavy logs
node generate_logs.js --scenario error --count 500

# Generate distributed trace logs
node generate_logs.js --scenario trace --count 100

# Generate at specific rate (100 logs/sec)
node generate_logs.js --rate 100 --count 1000

# Production-like distribution
node generate_logs.js --scenario production --count 5000
```

### Scenarios

| Scenario     | Mô tả                               |
| ------------ | ----------------------------------- |
| `random`     | Random level, random service        |
| `error`      | 60% ERROR, 30% FATAL, 10% WARN      |
| `debug`      | 70% DEBUG, 30% INFO (dev env)       |
| `production` | Realistic production distribution   |
| `trace`      | Distributed tracing (same trace_id) |

## Metrics

### k6 Metrics

- `http_req_duration`: Latency (p95, p99)
- `http_req_failed`: Error rate
- `logs_ingested`: Total logs sent
- `error_rate`: Custom error rate

### LogStorm Metrics

Check Prometheus: http://localhost:9090

- `logstorm_logs_processed_total`
- `logstorm_clickhouse_insert_duration_seconds`
- `logstorm_processor_buffer_size`

## Results

Results được lưu trong `./results/`:

```
results/
├── smoke_20260306_143022.json
├── load_20260306_144500.json
└── stress_20260306_151200.json
```

## Thresholds

| Metric      | Threshold |
| ----------- | --------- |
| p95 latency | < 500ms   |
| p99 latency | < 1000ms  |
| Error rate  | < 1%      |

## Tips

1. **Smoke test trước**: Chạy smoke test để verify hệ thống trước khi chạy load test
2. **Monitor resources**: Theo dõi CPU, Memory, Disk I/O khi chạy test
3. **Check Grafana**: Xem dashboard để phân tích kết quả
4. **Clean data**: Truncate ClickHouse tables giữa các test nếu cần

```sql
-- Clean test data
TRUNCATE TABLE logstorm.logs;
```
