#!/usr/bin/env bash
# Usage: bash scripts/seed-logs.sh <api-key>
# Sends a batch of sample logs to test the Logs page UI.

set -euo pipefail

API_KEY="${1:-}"
API_URL="${2:-http://localhost:8080/api/v1/logs}"

if [[ -z "$API_KEY" ]]; then
  echo "Usage: bash scripts/seed-logs.sh <api-key> [api-url]"
  echo "  api-key  — full API key from the UI (starts with lsk_)"
  echo "  api-url  — optional, default: http://localhost:8080/api/v1/logs"
  exit 1
fi

send() {
  local payload="$1"
  local res
  res=$(curl -sf -X POST "$API_URL" \
    -H "Content-Type: application/json" \
    -H "X-API-Key: $API_KEY" \
    -d "$payload" 2>&1) || {
    echo "  ERROR: $res"
    return
  }
  echo "  OK: $(echo "$res" | grep -o '"id":"[^"]*"' | head -1)"
}

echo "Sending logs to $API_URL ..."
echo ""

# --- payments-api ---
echo "[payments-api]"

send '{
  "level": "info",
  "service": "payments-api",
  "environment": "production",
  "message": "Payment processed successfully",
  "trace_id": "trace-pay-001",
  "span_id": "span-001",
  "attributes": {"payment_id": "pay_8kXz9", "amount": "149.99", "currency": "USD", "method": "card"}
}'

send '{
  "level": "warn",
  "service": "payments-api",
  "environment": "production",
  "message": "Payment retry attempt 2/3 — gateway timeout",
  "trace_id": "trace-pay-002",
  "span_id": "span-002",
  "attributes": {"payment_id": "pay_7mYq2", "attempt": "2", "gateway": "stripe"}
}'

send '{
  "level": "error",
  "service": "payments-api",
  "environment": "production",
  "message": "Payment failed: insufficient funds",
  "trace_id": "trace-pay-003",
  "span_id": "span-003",
  "attributes": {"payment_id": "pay_3nWr1", "error_code": "insufficient_funds", "amount": "299.00"}
}'

send '{
  "level": "debug",
  "service": "payments-api",
  "environment": "development",
  "message": "Webhook signature verified",
  "trace_id": "trace-pay-004",
  "attributes": {"event_type": "payment.succeeded", "webhook_id": "wh_Kp5zX"}
}'

send '{
  "level": "fatal",
  "service": "payments-api",
  "environment": "production",
  "message": "Database connection pool exhausted — all workers blocked",
  "attributes": {"pool_size": "10", "waiting": "47", "db_host": "pg-primary.internal"}
}'

echo ""

# --- auth-service ---
echo "[auth-service]"

send '{
  "level": "info",
  "service": "auth-service",
  "environment": "production",
  "message": "User logged in",
  "trace_id": "trace-auth-001",
  "attributes": {"user_id": "usr_Hj9kL", "ip": "203.0.113.42", "method": "password"}
}'

send '{
  "level": "warn",
  "service": "auth-service",
  "environment": "production",
  "message": "Failed login attempt — invalid password (3 of 5)",
  "trace_id": "trace-auth-002",
  "attributes": {"user_id": "usr_Hj9kL", "ip": "198.51.100.7", "attempt": "3"}
}'

send '{
  "level": "error",
  "service": "auth-service",
  "environment": "production",
  "message": "JWT signing key rotation failed — using cached key",
  "attributes": {"reason": "vault_unreachable", "key_age_hours": "26"}
}'

send '{
  "level": "info",
  "service": "auth-service",
  "environment": "staging",
  "message": "Refresh token issued",
  "trace_id": "trace-auth-003",
  "attributes": {"user_id": "usr_Mq3nP", "ttl_hours": "168"}
}'

send '{
  "level": "debug",
  "service": "auth-service",
  "environment": "development",
  "message": "OAuth2 callback received",
  "trace_id": "trace-auth-004",
  "attributes": {"provider": "google", "state_valid": "true"}
}'

echo ""

# --- notification-service ---
echo "[notification-service]"

send '{
  "level": "info",
  "service": "notification-service",
  "environment": "production",
  "message": "Email queued for delivery",
  "trace_id": "trace-notif-001",
  "attributes": {"template": "payment_receipt", "recipient": "u***@example.com", "queue_depth": "12"}
}'

send '{
  "level": "warn",
  "service": "notification-service",
  "environment": "production",
  "message": "Email delivery delayed — SES rate limit approaching (80%)",
  "attributes": {"sent_last_minute": "80", "limit": "100"}
}'

send '{
  "level": "error",
  "service": "notification-service",
  "environment": "production",
  "message": "SMS delivery failed after 3 retries",
  "trace_id": "trace-notif-002",
  "attributes": {"provider": "twilio", "error": "21211", "phone_country": "VN"}
}'

send '{
  "level": "info",
  "service": "notification-service",
  "environment": "production",
  "message": "Push notification sent to 1,204 devices",
  "attributes": {"campaign_id": "camp_Xz7", "platform": "fcm", "delivered": "1198", "failed": "6"}
}'

echo ""
echo "Done. Refresh the Logs page to see results."
