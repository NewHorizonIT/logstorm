import http from "k6/http";
import { check, sleep, group } from "k6";
import { Counter, Rate, Trend } from "k6/metrics";
import { randomString, randomIntBetween } from "https://jslib.k6.io/k6-utils/1.2.0/index.js";

// Custom metrics
const logsIngested = new Counter("logs_ingested");
const batchLogsIngested = new Counter("batch_logs_ingested");
const errorRate = new Rate("error_rate");
const ingestionDuration = new Trend("ingestion_duration_ms");

// Test configuration
const BASE_URL = __ENV.BASE_URL || "http://localhost:3123";

// Test scenarios
export const options = {
  scenarios: {
    // Scenario 1: Smoke test - low load
    smoke: {
      executor: "constant-vus",
      vus: 5,
      duration: "30s",
      tags: { scenario: "smoke" },
      exec: "singleLogTest",
    },
    // Scenario 2: Load test - normal traffic
    load: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "1m", target: 50 },
        { duration: "3m", target: 50 },
        { duration: "1m", target: 0 },
      ],
      tags: { scenario: "load" },
      exec: "mixedWorkload",
      startTime: "35s",
    },
    // Scenario 3: Stress test - high load
    stress: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "2m", target: 100 },
        { duration: "5m", target: 200 },
        { duration: "2m", target: 300 },
        { duration: "1m", target: 0 },
      ],
      tags: { scenario: "stress" },
      exec: "batchLogTest",
      startTime: "6m",
    },
    // Scenario 4: Spike test - sudden burst
    spike: {
      executor: "ramping-vus",
      startVUs: 0,
      stages: [
        { duration: "10s", target: 10 },
        { duration: "1m", target: 500 },
        { duration: "10s", target: 10 },
        { duration: "30s", target: 0 },
      ],
      tags: { scenario: "spike" },
      exec: "singleLogTest",
      startTime: "17m",
    },
  },
  thresholds: {
    http_req_duration: ["p(95)<500", "p(99)<1000"],
    http_req_failed: ["rate<0.01"],
    error_rate: ["rate<0.05"],
  },
};

// =============================================================================
// Log Generators
// =============================================================================

const SERVICES = [
  "user-service",
  "order-service",
  "payment-service",
  "inventory-service",
  "notification-service",
  "auth-service",
  "api-gateway",
  "search-service",
];

const ENVIRONMENTS = ["development", "staging", "production"];
const LEVELS = ["DEBUG", "INFO", "WARN", "ERROR", "FATAL"];
const LEVEL_WEIGHTS = [10, 50, 25, 12, 3]; // Distribution weights

function weightedRandom(items, weights) {
  const totalWeight = weights.reduce((a, b) => a + b, 0);
  let random = Math.random() * totalWeight;
  for (let i = 0; i < items.length; i++) {
    random -= weights[i];
    if (random <= 0) return items[i];
  }
  return items[items.length - 1];
}

function generateTraceId() {
  return `${randomString(8)}-${randomString(4)}-${randomString(4)}-${randomString(4)}-${randomString(12)}`;
}

// Generate realistic log messages based on level
function generateLogMessage(level, service) {
  const messages = {
    DEBUG: [
      `Processing request for ${service}`,
      `Cache hit for key: ${randomString(10)}`,
      `Database query executed in ${randomIntBetween(1, 100)}ms`,
      `Memory usage: ${randomIntBetween(50, 90)}%`,
      `Loaded configuration from ${randomString(8)}.yaml`,
    ],
    INFO: [
      `User ${randomString(8)} logged in successfully`,
      `Order ${randomString(10)} created`,
      `Payment processed: $${randomIntBetween(10, 1000)}.00`,
      `Request completed in ${randomIntBetween(10, 200)}ms`,
      `Connection established to ${service}`,
      `Health check passed`,
      `Scheduled job completed: ${randomString(12)}`,
    ],
    WARN: [
      `High memory usage detected: ${randomIntBetween(80, 95)}%`,
      `Slow query detected: ${randomIntBetween(500, 2000)}ms`,
      `Rate limit approaching for user ${randomString(8)}`,
      `Deprecated API called: /api/v1/${randomString(6)}`,
      `Connection pool nearly exhausted: ${randomIntBetween(80, 95)}%`,
      `Retry attempt ${randomIntBetween(1, 3)} for ${service}`,
    ],
    ERROR: [
      `Failed to connect to database: timeout after ${randomIntBetween(5, 30)}s`,
      `Invalid request payload: missing field '${randomString(6)}'`,
      `Authentication failed for user ${randomString(8)}`,
      `Service ${service} returned status ${randomIntBetween(500, 503)}`,
      `Transaction rollback: ${randomString(16)}`,
      `File not found: /data/${randomString(10)}.json`,
    ],
    FATAL: [
      `Out of memory: available ${randomIntBetween(0, 5)}MB`,
      `Database connection pool exhausted`,
      `Critical service ${service} unreachable`,
      `Disk space critical: ${randomIntBetween(95, 99)}% used`,
      `Unrecoverable error in ${service}: ${randomString(20)}`,
    ],
  };

  const levelMessages = messages[level] || messages.INFO;
  return levelMessages[Math.floor(Math.random() * levelMessages.length)];
}

// Generate a single log entry
function generateLog() {
  const service = SERVICES[Math.floor(Math.random() * SERVICES.length)];
  const level = weightedRandom(LEVELS, LEVEL_WEIGHTS);
  const environment = ENVIRONMENTS[Math.floor(Math.random() * ENVIRONMENTS.length)];

  return {
    id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER),
    message: generateLogMessage(level, service),
    trace_id: generateTraceId(),
    environment: environment,
    level: level,
    service: service,
    timestamp: Date.now(),
  };
}

// Generate batch of logs
function generateLogBatch(size) {
  const logs = [];
  // Use same trace_id for some logs to simulate distributed tracing
  const sharedTraceId = generateTraceId();
  const useSharedTrace = Math.random() > 0.7;

  for (let i = 0; i < size; i++) {
    const log = generateLog();
    if (useSharedTrace && i < size / 2) {
      log.trace_id = sharedTraceId;
    }
    logs.push(log);
  }
  return logs;
}

// =============================================================================
// Test Functions
// =============================================================================

const headers = { "Content-Type": "application/json" };

// Single log ingestion test
export function singleLogTest() {
  const log = generateLog();
  const payload = JSON.stringify(log);

  const startTime = Date.now();
  const res = http.post(`${BASE_URL}/ingestion/logs`, payload, { headers });
  const duration = Date.now() - startTime;

  ingestionDuration.add(duration);
  logsIngested.add(1);

  const success = check(res, {
    "status is 200": (r) => r.status === 200,
    "response has message": (r) => r.json("message") !== undefined,
  });

  errorRate.add(!success);
  sleep(randomIntBetween(50, 200) / 1000);
}

// Batch log ingestion test
export function batchLogTest() {
  const batchSize = randomIntBetween(10, 100);
  const logs = generateLogBatch(batchSize);

  // Send each log in the batch
  for (const log of logs) {
    const payload = JSON.stringify(log);
    const res = http.post(`${BASE_URL}/ingestion/logs`, payload, { headers });

    const success = check(res, {
      "batch log status is 200": (r) => r.status === 200,
    });

    errorRate.add(!success);
    batchLogsIngested.add(1);
  }

  sleep(randomIntBetween(100, 500) / 1000);
}

// Mixed workload - simulates real traffic patterns
export function mixedWorkload() {
  group("mixed_workload", function () {
    // 70% single logs
    if (Math.random() < 0.7) {
      singleLogTest();
    } else {
      // 30% small batches
      const batchSize = randomIntBetween(5, 20);
      const logs = generateLogBatch(batchSize);

      for (const log of logs) {
        const payload = JSON.stringify(log);
        const res = http.post(`${BASE_URL}/ingestion/logs`, payload, { headers });
        logsIngested.add(1);

        check(res, {
          "mixed batch status is 200": (r) => r.status === 200,
        });
      }
    }
  });

  sleep(randomIntBetween(50, 300) / 1000);
}

// Health check
export function healthCheck() {
  const res = http.get(`${BASE_URL}/health`);
  check(res, {
    "health check status is 200": (r) => r.status === 200,
  });
}

// =============================================================================
// Lifecycle Hooks
// =============================================================================

export function setup() {
  console.log("Starting LogStorm Performance Test");
  console.log(`Target: ${BASE_URL}`);

  // Verify service is up
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200) {
    throw new Error(`Service not available at ${BASE_URL}`);
  }

  return { startTime: Date.now() };
}

export function teardown(data) {
  const duration = (Date.now() - data.startTime) / 1000;
  console.log(`Test completed in ${duration.toFixed(2)}s`);
}

// Default function for quick runs
export default function () {
  singleLogTest();
}