import http from "k6/http";
import { check, sleep } from "k6";
import { Counter, Rate } from "k6/metrics";

// Spike test - sudden burst of traffic
const logsIngested = new Counter("logs_ingested");
const errorRate = new Rate("error_rate");

const BASE_URL = __ENV.BASE_URL || "http://localhost:3123";

export const options = {
  stages: [
    { duration: "30s", target: 10 },    // Normal load
    { duration: "10s", target: 1000 },  // SPIKE!
    { duration: "1m", target: 1000 },   // Stay at spike
    { duration: "10s", target: 10 },    // Back to normal
    { duration: "30s", target: 10 },    // Recovery period
    { duration: "10s", target: 0 },     // Ramp down
  ],
  thresholds: {
    http_req_failed: ["rate<0.2"],  // Allow some errors during spike
  },
};

const SERVICES = ["api-gateway", "auth-service", "user-service"];

function randomString(len) {
  return Math.random().toString(36).substring(2, 2 + len);
}

export default function () {
  const log = {
    id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER),
    message: `Spike test log: ${randomString(20)}`,
    trace_id: `spike-${randomString(8)}-${randomString(4)}-${randomString(12)}`,
    environment: "production",
    level: "INFO",
    service: SERVICES[Math.floor(Math.random() * SERVICES.length)],
    timestamp: Date.now(),
  };

  const res = http.post(`${BASE_URL}/ingestion/logs`, JSON.stringify(log), {
    headers: { "Content-Type": "application/json" },
  });

  const success = check(res, {
    "status is 200 or 429": (r) => r.status === 200 || r.status === 429,
  });

  logsIngested.add(1);
  errorRate.add(!success);

  // No sleep during spike to maximize pressure
}
