#!/usr/bin/env bash
# smoke-auth.sh — end-to-end smoke test for auth domain (Phase 2)
#
# Prerequisites:
#   1. docker compose up -d
#   2. ./scripts/seed.sh          (creates demo tenant + user + api key)
#   3. make run                   (API server on :8000)
#
# Usage:
#   ./scripts/smoke-auth.sh
#   API_BASE=http://localhost:8000 ./scripts/smoke-auth.sh
set -euo pipefail

: "${API_BASE:=http://localhost:8000}"

RED='\033[0;31m'
GREEN='\033[0;32m'
NC='\033[0m' # No Color

pass() { echo -e "${GREEN}✓ $1${NC}"; }
fail() { echo -e "${RED}✗ $1${NC}"; exit 1; }

# ── Step 1: Login ────────────────────────────────────────────────────────────
echo "Step 1: POST /v1/auth/login"
LOGIN_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/v1/auth/login" \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"demo1234"}')

LOGIN_BODY=$(echo "$LOGIN_RESP" | head -n -1)
LOGIN_STATUS=$(echo "$LOGIN_RESP" | tail -n 1)

if [ "$LOGIN_STATUS" != "200" ]; then
  fail "Login returned $LOGIN_STATUS: $LOGIN_BODY"
fi

ACCESS_TOKEN=$(echo "$LOGIN_BODY" | jq -r '.data.access_token')
REFRESH_TOKEN=$(echo "$LOGIN_BODY" | jq -r '.data.refresh_token')

if [ -z "$ACCESS_TOKEN" ] || [ "$ACCESS_TOKEN" = "null" ]; then
  fail "Login response missing access_token"
fi
if [ -z "$REFRESH_TOKEN" ] || [ "$REFRESH_TOKEN" = "null" ]; then
  fail "Login response missing refresh_token"
fi
pass "Login succeeded (access_token length=${#ACCESS_TOKEN})"

# ── Step 2: Refresh ──────────────────────────────────────────────────────────
echo "Step 2: POST /v1/auth/refresh"
REFRESH_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/v1/auth/refresh" \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}")

REFRESH_BODY=$(echo "$REFRESH_RESP" | head -n -1)
REFRESH_STATUS=$(echo "$REFRESH_RESP" | tail -n 1)

if [ "$REFRESH_STATUS" != "200" ]; then
  fail "Refresh returned $REFRESH_STATUS: $REFRESH_BODY"
fi

NEW_ACCESS=$(echo "$REFRESH_BODY" | jq -r '.data.access_token')
if [ -z "$NEW_ACCESS" ] || [ "$NEW_ACCESS" = "null" ]; then
  fail "Refresh response missing access_token"
fi
# Use the new access token for subsequent requests.
ACCESS_TOKEN="$NEW_ACCESS"
pass "Token refresh succeeded"

# ── Step 3: Create API Key ───────────────────────────────────────────────────
echo "Step 3: POST /v1/auth/keys"
KEY_RESP=$(curl -s -w "\n%{http_code}" -X POST "$API_BASE/v1/auth/keys" \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"name":"smoke-test-key"}')

KEY_BODY=$(echo "$KEY_RESP" | head -n -1)
KEY_STATUS=$(echo "$KEY_RESP" | tail -n 1)

if [ "$KEY_STATUS" != "200" ]; then
  fail "Create API key returned $KEY_STATUS: $KEY_BODY"
fi

RAW_KEY=$(echo "$KEY_BODY" | jq -r '.data.raw_key')
KEY_ID=$(echo "$KEY_BODY" | jq -r '.data.id')

if [ -z "$RAW_KEY" ] || [ "$RAW_KEY" = "null" ]; then
  fail "Create API key response missing raw_key"
fi
if [ -z "$KEY_ID" ] || [ "$KEY_ID" = "null" ]; then
  fail "Create API key response missing id"
fi
pass "API key created (id=$KEY_ID, raw_key prefix=${RAW_KEY:0:6}...)"

# ── Step 4: List API Keys ───────────────────────────────────────────────────
echo "Step 4: GET /v1/auth/keys"
LIST_RESP=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/v1/auth/keys" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

LIST_BODY=$(echo "$LIST_RESP" | head -n -1)
LIST_STATUS=$(echo "$LIST_RESP" | tail -n 1)

if [ "$LIST_STATUS" != "200" ]; then
  fail "List API keys returned $LIST_STATUS: $LIST_BODY"
fi

KEY_COUNT=$(echo "$LIST_BODY" | jq '.data.keys | length')
if [ "$KEY_COUNT" -lt 1 ]; then
  fail "Expected at least 1 key, got $KEY_COUNT"
fi
pass "Listed $KEY_COUNT API key(s)"

# ── Step 5: Revoke API Key ──────────────────────────────────────────────────
echo "Step 5: DELETE /v1/auth/keys/$KEY_ID"
DEL_RESP=$(curl -s -w "\n%{http_code}" -X DELETE "$API_BASE/v1/auth/keys/$KEY_ID" \
  -H "Authorization: Bearer $ACCESS_TOKEN")

DEL_BODY=$(echo "$DEL_RESP" | head -n -1)
DEL_STATUS=$(echo "$DEL_RESP" | tail -n 1)

if [ "$DEL_STATUS" != "200" ]; then
  fail "Revoke API key returned $DEL_STATUS: $DEL_BODY"
fi
pass "API key $KEY_ID revoked"

# ── Step 6: Verify 401 without token ────────────────────────────────────────
echo "Step 6: GET /v1/auth/keys (no auth — expect 401)"
NOAUTH_RESP=$(curl -s -w "\n%{http_code}" -X GET "$API_BASE/v1/auth/keys")

NOAUTH_STATUS=$(echo "$NOAUTH_RESP" | tail -n 1)

if [ "$NOAUTH_STATUS" != "401" ]; then
  fail "Expected 401 without auth, got $NOAUTH_STATUS"
fi
pass "Protected endpoint returned 401 without token"

echo ""
echo -e "${GREEN}All smoke tests passed!${NC}"
