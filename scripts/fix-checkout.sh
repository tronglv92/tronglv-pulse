#!/usr/bin/env bash
# fix-checkout.sh — revert checkout back to 1% error rate
set -euo pipefail

: "${TARGET:=http://localhost:8002}"
: "${API_KEY:=demo-key}"
: "${HMAC_SECRET:=demo-secret}"

go run ./cmd/sim/main.go \
  --target "$TARGET" \
  --api-key "$API_KEY" \
  --hmac-secret "$HMAC_SECRET" \
  --service checkout-svc \
  --rps 100 \
  --error-rate 0.01 \
  --duration 60s
