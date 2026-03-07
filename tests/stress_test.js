import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Trend } from "k6/metrics";

// Stress test - find the breaking point
const logsIngested = new Counter("logs_ingested");
const latency = new Trend("latency_ms");

const BASE_URL = __ENV.BASE_URL || "http://localhost:3123";

export const options = {
  stages: [
    { duration: "2m", target: 100 },   // Ramp up to 100 VUs
    { duration: "5m", target: 100 },   // Stay at 100 VUs
    { duration: "2m", target: 200 },   // Ramp up to 200 VUs
    { duration: "5m", target: 200 },   // Stay at 200 VUs
    { duration: "2m", target: 300 },   // Ramp up to 300 VUs
    { duration: "5m", target: 300 },   // Stay at 300 VUs
    { duration: "2m", target: 500 },   // Ramp up to 500 VUs
    { duration: "5m", target: 500 },   // Stay at 500 VUs
    { duration: "5m", target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_failed: ["rate<0.1"],     // Allow higher error rate for stress test
    latency_ms: ["p(95)<2000"],
  },
};

const SERVICES = [
  "user-service", "order-service", "payment-service",
  "inventory-service", "notification-service", "auth-service",
];
const LEVELS = ["DEBUG", "INFO", "WARN", "ERROR", "FATAL"];
const LEVEL_WEIGHTS = [5, 60, 20, 12, 3];

function weightedRandom(items, weights) {
  const total = weights.reduce((a, b) => a + b, 0);
  let r = Math.random() * total;
  for (let i = 0; i < items.length; i++) {
    r -= weights[i];
    if (r <= 0) return items[i];
  }
  return items[items.length - 1];
}

function randomString(len) {
  return Math.random().toString(36).substring(2, 2 + len);
}

function generateLog() {
  const service = SERVICES[Math.floor(Math.random() * SERVICES.length)];
  const level = weightedRandom(LEVELS, LEVEL_WEIGHTS);

  const messages = {
    DEBUG: `Debug info: ${randomString(30)}`,
    INFO: `Request processed successfully for user ${randomString(8)}`,
    WARN: `High latency detected: ${Math.floor(Math.random() * 1000)}ms`,
    ERROR: `Failed to process request: ${randomString(20)}`,
    FATAL: `Critical error in ${service}: ${randomString(15)}`,
  };

  return {
    id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER),
    message: messages[level],
    trace_id: `${randomString(8)}-${randomString(4)}-${randomString(4)}-${randomString(12)}`,
    environment: ["development", "staging", "production"][Math.floor(Math.random() * 3)],
    level: level,
    service: service,
    timestamp: Date.now(),
  };
}

export default function () {
  const log = generateLog();
  const start = Date.now();

  const res = http.post(`${BASE_URL}/ingestion/logs`, JSON.stringify(log), {
    headers: { "Content-Type": "application/json" },
  });

  latency.add(Date.now() - start);
  logsIngested.add(1);

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  // Minimal sleep to maximize throughput
  sleep(0.01);
}
