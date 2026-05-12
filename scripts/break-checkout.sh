#!/usr/bin/env bash
# break-checkout.sh — simulate checkout outage at 50% error rate
set -euo pipefail

: "${TARGET:=http://localhost:8002}"
: "${API_KEY:=demo-key}"
: "${HMAC_SECRET:=demo-secret}"

go run ./cmd/sim/main.go \
  --target "$TARGET" \
  --api-key "$API_KEY" \
  --hmac-secret "$HMAC_SECRET" \
  --service checkout-svc \
  --rps 50 \
  --error-rate 0.50 \
  --duration 120s
