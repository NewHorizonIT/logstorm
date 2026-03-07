import http from "k6/http";
import { check, sleep } from "k6";
import { Counter } from "k6/metrics";

// Soak test - long running test to detect memory leaks
const logsIngested = new Counter("logs_ingested");

const BASE_URL = __ENV.BASE_URL || "http://localhost:3123";

export const options = {
  stages: [
    { duration: "2m", target: 50 },     // Ramp up
    { duration: "4h", target: 50 },     // Stay at 50 VUs for 4 hours
    { duration: "2m", target: 0 },      // Ramp down
  ],
  thresholds: {
    http_req_duration: ["p(99)<1000"],
    http_req_failed: ["rate<0.01"],
  },
};

const SERVICES = [
  "user-service", "order-service", "payment-service",
  "inventory-service", "notification-service",
];
const LEVELS = ["DEBUG", "INFO", "WARN", "ERROR"];
const ENVS = ["production"];  // Focus on production logs

function randomString(len) {
  return Math.random().toString(36).substring(2, 2 + len);
}

function generateRealisticLog() {
  const service = SERVICES[Math.floor(Math.random() * SERVICES.length)];
  const level = LEVELS[Math.floor(Math.random() * LEVELS.length)];

  // Simulate real-world log patterns
  const patterns = {
    "user-service": [
      "User login successful",
      "User profile updated",
      "Password change requested",
      "Session expired",
    ],
    "order-service": [
      "Order created",
      "Order status updated to PROCESSING",
      "Order shipped",
      "Order delivered",
    ],
    "payment-service": [
      "Payment initiated",
      "Payment authorized",
      "Payment captured",
      "Refund processed",
    ],
    "inventory-service": [
      "Stock updated",
      "Low stock alert",
      "Item reserved",
      "Reservation expired",
    ],
    "notification-service": [
      "Email sent",
      "SMS queued",
      "Push notification delivered",
      "Notification failed - retry scheduled",
    ],
  };

  const servicePatterns = patterns[service] || ["Generic log message"];
  const message = servicePatterns[Math.floor(Math.random() * servicePatterns.length)];

  return {
    id: Math.floor(Math.random() * Number.MAX_SAFE_INTEGER),
    message: `${message} [${randomString(8)}]`,
    trace_id: `${randomString(8)}-${randomString(4)}-${randomString(4)}-${randomString(12)}`,
    environment: "production",
    level: level,
    service: service,
    timestamp: Date.now(),
  };
}

export default function () {
  const log = generateRealisticLog();

  const res = http.post(`${BASE_URL}/ingestion/logs`, JSON.stringify(log), {
    headers: { "Content-Type": "application/json" },
  });

  check(res, {
    "status is 200": (r) => r.status === 200,
  });

  logsIngested.add(1);

  // Realistic delay between logs
  sleep(Math.random() * 0.2 + 0.1);
}
