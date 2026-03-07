#!/usr/bin/env node

/**
 * Log Generator for LogStorm
 * 
 * Usage:
 *   node generate_logs.js                    # Generate 100 random logs
 *   node generate_logs.js --count 1000       # Generate 1000 logs
 *   node generate_logs.js --scenario error   # Generate error-heavy logs
 *   node generate_logs.js --scenario trace   # Generate distributed trace logs
 *   node generate_logs.js --rate 100         # Generate at 100 logs/sec
 *   node generate_logs.js --duration 60      # Run for 60 seconds
 */

const http = require("http");

const BASE_URL = process.env.BASE_URL || "http://localhost:3123";
const [, , ...args] = process.argv;

// Parse arguments
function parseArgs(args) {
  const options = {
    count: 100,
    scenario: "random",
    rate: 0,        // 0 = as fast as possible
    duration: 0,    // 0 = until count reached
    service: null,
    level: null,
  };

  for (let i = 0; i < args.length; i += 2) {
    const key = args[i].replace("--", "");
    const value = args[i + 1];
    if (key in options) {
      options[key] = isNaN(value) ? value : parseInt(value);
    }
  }
  return options;
}

// Services
const SERVICES = [
  "user-service",
  "order-service",
  "payment-service",
  "inventory-service",
  "notification-service",
  "auth-service",
  "api-gateway",
  "search-service",
  "analytics-service",
  "recommendation-service",
];

const ENVIRONMENTS = ["development", "staging", "production"];
const LEVELS = ["DEBUG", "INFO", "WARN", "ERROR", "FATAL"];

// Log message templates per service and level
const MESSAGE_TEMPLATES = {
  "user-service": {
    DEBUG: [
      "Fetching user profile for userId={userId}",
      "Cache lookup for user session: {sessionId}",
      "Validating user permissions",
    ],
    INFO: [
      "User {userId} logged in successfully",
      "New user registered: {email}",
      "Password updated for user {userId}",
      "User profile updated",
    ],
    WARN: [
      "Multiple failed login attempts for user {userId}",
      "Session about to expire for {userId}",
      "User quota almost reached: {percentage}%",
    ],
    ERROR: [
      "Failed to authenticate user {userId}: invalid credentials",
      "Database connection timeout while fetching user",
      "Email verification failed for {email}",
    ],
    FATAL: [
      "User database unreachable",
      "Authentication service crashed",
    ],
  },
  "order-service": {
    DEBUG: [
      "Processing order {orderId}",
      "Calculating shipping cost for {destination}",
      "Validating order items",
    ],
    INFO: [
      "Order {orderId} created successfully",
      "Order {orderId} status changed to {status}",
      "Refund processed for order {orderId}",
    ],
    WARN: [
      "Order {orderId} stuck in processing for {minutes} minutes",
      "High order volume detected: {count} orders in queue",
    ],
    ERROR: [
      "Failed to create order: {reason}",
      "Payment validation failed for order {orderId}",
      "Inventory check failed for order {orderId}",
    ],
    FATAL: [
      "Order processing pipeline stopped",
      "Critical: Unable to connect to payment gateway",
    ],
  },
  "payment-service": {
    DEBUG: [
      "Initiating payment for {amount}",
      "Verifying card details",
      "Checking fraud score",
    ],
    INFO: [
      "Payment {paymentId} authorized",
      "Payment {paymentId} captured: ${amount}",
      "Refund {refundId} processed",
    ],
    WARN: [
      "High fraud score detected: {score}",
      "Payment retry scheduled: attempt {attempt}",
      "Rate limit approaching for merchant {merchantId}",
    ],
    ERROR: [
      "Payment declined: {reason}",
      "Card verification failed",
      "Timeout connecting to payment provider",
    ],
    FATAL: [
      "Payment gateway connection lost",
      "Critical: Encryption key unavailable",
    ],
  },
};

// Generate random values
function randomString(len) {
  return Math.random().toString(36).substring(2, 2 + len);
}

function randomInt(min, max) {
  return Math.floor(Math.random() * (max - min + 1)) + min;
}

function randomItem(arr) {
  return arr[Math.floor(Math.random() * arr.length)];
}

function generateTraceId() {
  return `${randomString(8)}-${randomString(4)}-4${randomString(3)}-${randomString(4)}-${randomString(12)}`;
}

// Template replacements
function fillTemplate(template) {
  return template
    .replace("{userId}", `usr_${randomString(8)}`)
    .replace("{sessionId}", `sess_${randomString(12)}`)
    .replace("{email}", `user_${randomString(6)}@example.com`)
    .replace("{orderId}", `ord_${randomString(10)}`)
    .replace("{paymentId}", `pay_${randomString(10)}`)
    .replace("{refundId}", `ref_${randomString(10)}`)
    .replace("{merchantId}", `mer_${randomString(8)}`)
    .replace("{amount}", randomInt(10, 999).toString())
    .replace("{percentage}", randomInt(80, 99).toString())
    .replace("{minutes}", randomInt(5, 30).toString())
    .replace("{count}", randomInt(100, 1000).toString())
    .replace("{score}", randomInt(70, 99).toString())
    .replace("{attempt}", randomInt(1, 3).toString())
    .replace("{status}", randomItem(["PROCESSING", "SHIPPED", "DELIVERED"]))
    .replace("{reason}", randomItem(["insufficient_funds", "invalid_card", "expired_card"]))
    .replace("{destination}", randomItem(["US", "UK", "DE", "JP", "AU"]));
}

// Generate log based on scenario
function generateLog(scenario, options = {}) {
  const service = options.service || randomItem(SERVICES);
  let level;
  let environment = randomItem(ENVIRONMENTS);

  switch (scenario) {
    case "error":
      // 60% errors, 30% fatal, 10% warn
      level = Math.random() < 0.6 ? "ERROR" : Math.random() < 0.75 ? "FATAL" : "WARN";
      environment = "production";
      break;

    case "debug":
      // 70% debug, 30% info
      level = Math.random() < 0.7 ? "DEBUG" : "INFO";
      environment = "development";
      break;

    case "production":
      // Production-like distribution
      const r = Math.random();
      if (r < 0.4) level = "INFO";
      else if (r < 0.7) level = "DEBUG";
      else if (r < 0.85) level = "WARN";
      else if (r < 0.97) level = "ERROR";
      else level = "FATAL";
      environment = "production";
      break;

    default:
      level = options.level || randomItem(LEVELS);
  }

  // Get message template
  const templates = MESSAGE_TEMPLATES[service]?.[level] || [`${level} log from ${service}: ${randomString(20)}`];
  const message = fillTemplate(randomItem(templates));

  return {
    id: randomInt(1, Number.MAX_SAFE_INTEGER),
    message: message,
    trace_id: generateTraceId(),
    environment: environment,
    level: level,
    service: service,
    timestamp: Date.now(),
  };
}

// Generate distributed trace logs (multiple services, same trace)
function generateTraceLogs(traceId = null) {
  traceId = traceId || generateTraceId();
  const logs = [];

  // Simulate a request flowing through services
  const flow = [
    { service: "api-gateway", level: "INFO", msg: "Incoming request received" },
    { service: "auth-service", level: "DEBUG", msg: "Validating JWT token" },
    { service: "auth-service", level: "INFO", msg: "Token validated successfully" },
    { service: "user-service", level: "DEBUG", msg: "Fetching user context" },
    { service: "order-service", level: "INFO", msg: "Creating new order" },
    { service: "inventory-service", level: "DEBUG", msg: "Checking item availability" },
    { service: "inventory-service", level: "INFO", msg: "Items reserved" },
    { service: "payment-service", level: "INFO", msg: "Payment initiated" },
    { service: "payment-service", level: "INFO", msg: "Payment authorized" },
    { service: "notification-service", level: "INFO", msg: "Order confirmation email queued" },
    { service: "api-gateway", level: "INFO", msg: "Request completed successfully" },
  ];

  let timestamp = Date.now();
  for (const step of flow) {
    logs.push({
      id: randomInt(1, Number.MAX_SAFE_INTEGER),
      message: step.msg,
      trace_id: traceId,
      environment: "production",
      level: step.level,
      service: step.service,
      timestamp: timestamp,
    });
    timestamp += randomInt(5, 50); // Small delay between services
  }

  return logs;
}

// Send log to LogStorm
async function sendLog(log) {
  return new Promise((resolve, reject) => {
    const url = new URL(`${BASE_URL}/ingestion/logs`);
    const data = JSON.stringify(log);

    const options = {
      hostname: url.hostname,
      port: url.port || 80,
      path: url.pathname,
      method: "POST",
      headers: {
        "Content-Type": "application/json",
        "Content-Length": Buffer.byteLength(data),
      },
    };

    const req = http.request(options, (res) => {
      let body = "";
      res.on("data", (chunk) => (body += chunk));
      res.on("end", () => resolve({ status: res.statusCode, body }));
    });

    req.on("error", reject);
    req.write(data);
    req.end();
  });
}

// Main runner
async function main() {
  const options = parseArgs(args);
  console.log("LogStorm Log Generator");
  console.log("======================");
  console.log(`Target: ${BASE_URL}`);
  console.log(`Scenario: ${options.scenario}`);
  console.log(`Count: ${options.count}`);
  if (options.rate > 0) console.log(`Rate: ${options.rate} logs/sec`);
  console.log("");

  let sent = 0;
  let errors = 0;
  const startTime = Date.now();
  const delay = options.rate > 0 ? 1000 / options.rate : 0;

  // Handle trace scenario differently
  if (options.scenario === "trace") {
    const traceCount = Math.ceil(options.count / 11); // ~11 logs per trace
    for (let i = 0; i < traceCount; i++) {
      const traceLogs = generateTraceLogs();
      for (const log of traceLogs) {
        try {
          await sendLog(log);
          sent++;
          process.stdout.write(`\rSent: ${sent} | Errors: ${errors}`);
        } catch (e) {
          errors++;
        }
        if (delay > 0) await sleep(delay);
      }
    }
  } else {
    // Regular scenarios
    for (let i = 0; i < options.count; i++) {
      const log = generateLog(options.scenario, options);
      try {
        await sendLog(log);
        sent++;
        process.stdout.write(`\rSent: ${sent} | Errors: ${errors}`);
      } catch (e) {
        errors++;
      }
      if (delay > 0) await sleep(delay);
    }
  }

  const elapsed = (Date.now() - startTime) / 1000;
  console.log("\n");
  console.log("Summary");
  console.log("-------");
  console.log(`Total sent: ${sent}`);
  console.log(`Errors: ${errors}`);
  console.log(`Duration: ${elapsed.toFixed(2)}s`);
  console.log(`Rate: ${(sent / elapsed).toFixed(2)} logs/sec`);
}

function sleep(ms) {
  return new Promise((resolve) => setTimeout(resolve, ms));
}

main().catch(console.error);
