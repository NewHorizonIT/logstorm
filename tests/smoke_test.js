import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

// Quick smoke test - run this first to verify everything works
const errorRate = new Rate("error_rate");
const logsIngested = new Counter("logs_ingested");

const BASE_URL = __ENV.BASE_URL || "http://localhost:3123";

export const options = {
  vus: 10,
  duration: "1m",
  thresholds: {
    http_req_duration: ["p(95)<200"],
    http_req_failed: ["rate<0.01"],
  },
};

const SERVICES = ["user-service", "order-service", "payment-service"];
const LEVELS = ["DEBUG", "INFO", "WARN", "ERROR"];
const ENVS = ["development", "staging", "production"];

function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function randomString(len) {
  return Math.random().toString(36).substring(2, 2 + len);
}

export default function () {
  const log = {
    id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER),
    message: `Test message from ${randomItem(SERVICES)}: ${randomString(20)}`,
    trace_id: `${randomString(8)}-${randomString(4)}-${randomString(4)}-${randomString(12)}`,
    environment: randomItem(ENVS),
    level: randomItem(LEVELS),
    service: randomItem(SERVICES),
    timestamp: Date.now(),
  };

  const res = http.post(`${BASE_URL}/ingestion/logs`, JSON.stringify(log), {
    headers: { "Content-Type": "application/json" },
  });

  const success = check(res, {
    "status is 200": (r) => r.status === 200,
  });

  logsIngested.add(1);
  errorRate.add(!success);

  sleep(0.05);
}
