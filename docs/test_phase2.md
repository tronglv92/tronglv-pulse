# Phase 2 Test Plan — `auth` Domain

## Prerequisites

```bash
# 1. Start local infrastructure
docker compose up -d

# 2. Seed demo data (tenant, user, API key)
./scripts/seed.sh

# 3. Start the API server
make run
# or: go run cmd/api/main.go -f etc/api.yaml
```

**Seeded demo credentials:**

| Field | Value |
|---|---|
| Email | `demo@pulse.dev` |
| Password | `demo1234` |
| Tenant slug | `demo` |
| HMAC secret | `hmac_demo_secret_change_me` |
| API key (raw) | `pk_demo_0000000000000000` |

**Base URL:** `http://localhost:8000`

**Response envelope format:** All responses use `{"meta": {...}, "data": ...}`.

---

## Part 1: Manual Testing (curl)

### 1.1 Login — Happy Path

```bash
curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"demo1234"}' | jq .
```

**Expected:** HTTP 200

```json
{
  "meta": { "code": 200 },
  "data": {
    "access_token": "<jwt-string>",
    "refresh_token": "<jwt-string>",
    "expires_in": 900
  }
}
```

**Verify:**
- `expires_in` is `900` (15 minutes in seconds)
- Both tokens are valid JWT strings (three dot-separated base64 segments)
- Decode access token at https://jwt.io — claims contain `"eid"`, `"kind":"user"`, `"iss":"pulse"`, `"attributes":{"tenant_id":"<id>","email":"demo@pulse.dev"}`
- Decode refresh token — claims contain `"kind":"refresh"`

Save the tokens for subsequent steps:
```bash
ACCESS_TOKEN=$(curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"demo1234"}' | jq -r '.data.access_token')

REFRESH_TOKEN=$(curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"demo1234"}' | jq -r '.data.refresh_token')
```

---

### 1.2 Login — Wrong Password

```bash
curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"wrongpassword"}' | jq .
```

**Expected:** HTTP 401, reason `INVALID_CREDENTIALS`

---

### 1.3 Login — Non-existent Email

```bash
curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"nobody@pulse.dev","password":"demo1234"}' | jq .
```

**Expected:** HTTP 401, reason `INVALID_CREDENTIALS`

---

### 1.4 Login — Missing Fields

```bash
# Empty body
curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Expected:** HTTP 400, reason `INVALID_INPUT`

```bash
# Missing password
curl -s -X POST http://localhost:8000/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev"}' | jq .
```

**Expected:** HTTP 400, reason `INVALID_INPUT`

```bash
# No JSON body at all
curl -s -X POST http://localhost:8000/v1/auth/login | jq .
```

**Expected:** HTTP 400, reason `INVALID_BODY`

---

### 1.5 Refresh — Happy Path

```bash
curl -s -X POST http://localhost:8000/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}" | jq .
```

**Expected:** HTTP 200, new `access_token` and `refresh_token` returned, `expires_in` is `900`.

---

### 1.6 Refresh — Using Access Token (should fail)

```bash
curl -s -X POST http://localhost:8000/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$ACCESS_TOKEN\"}" | jq .
```

**Expected:** HTTP 401, reason `INVALID_TOKEN_KIND`

---

### 1.7 Refresh — Expired / Garbage Token

```bash
curl -s -X POST http://localhost:8000/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{"refresh_token":"not.a.valid.jwt"}' | jq .
```

**Expected:** HTTP 401, reason `INVALID_TOKEN`

---

### 1.8 Refresh — Missing Field

```bash
curl -s -X POST http://localhost:8000/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d '{}' | jq .
```

**Expected:** HTTP 400, reason `INVALID_INPUT`

---

### 1.9 Create API Key — Happy Path

```bash
curl -s -X POST http://localhost:8000/v1/auth/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"name":"my-test-key"}' | jq .
```

**Expected:** HTTP 200

```json
{
  "meta": { "code": 200 },
  "data": {
    "id": 2,
    "name": "my-test-key",
    "raw_key": "pk_<48-hex-characters>",
    "created_at": "...",
    "expires_at": null
  }
}
```

**Verify:**
- `raw_key` starts with `pk_` and is 51 characters total (`pk_` + 48 hex)
- `id` is a positive integer
- Save `KEY_ID` and `RAW_KEY` for next steps:
```bash
RESP=$(curl -s -X POST http://localhost:8000/v1/auth/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"name":"my-test-key"}')
KEY_ID=$(echo "$RESP" | jq -r '.data.id')
RAW_KEY=$(echo "$RESP" | jq -r '.data.raw_key')
echo "KEY_ID=$KEY_ID  RAW_KEY=$RAW_KEY"
```

---

### 1.10 Create API Key — With Expiry

```bash
curl -s -X POST http://localhost:8000/v1/auth/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{"name":"expiring-key","expires_at":"2026-12-31T23:59:59Z"}' | jq .
```

**Expected:** HTTP 200, `expires_at` field is set in the response.

---

### 1.11 Create API Key — Missing Name

```bash
curl -s -X POST http://localhost:8000/v1/auth/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -d '{}' | jq .
```

**Expected:** HTTP 400, reason `INVALID_INPUT`

---

### 1.12 Create API Key — No Auth

```bash
curl -s -X POST http://localhost:8000/v1/auth/keys \
  -H "Content-Type: application/json" \
  -d '{"name":"sneaky-key"}' | jq .
```

**Expected:** HTTP 401, reason `MISSING_TOKEN`

---

### 1.13 List API Keys — Happy Path

```bash
curl -s -X GET http://localhost:8000/v1/auth/keys \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Expected:** HTTP 200

```json
{
  "meta": { "code": 200 },
  "data": {
    "keys": [
      {
        "id": 2,
        "name": "my-test-key",
        "created_at": "...",
        "expires_at": null,
        "last_used_at": null
      }
    ]
  }
}
```

**Verify:**
- Response does NOT contain `raw_key` or `key_hash` fields
- The key created in step 1.9 appears in the list
- `last_used_at` is null for newly created keys

---

### 1.14 List API Keys — No Auth

```bash
curl -s -X GET http://localhost:8000/v1/auth/keys | jq .
```

**Expected:** HTTP 401, reason `MISSING_TOKEN`

---

### 1.15 Revoke API Key — Happy Path

```bash
curl -s -X DELETE http://localhost:8000/v1/auth/keys/$KEY_ID \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Expected:** HTTP 200 (success message, no data body)

**Verify:** List keys again — the revoked key should no longer appear:
```bash
curl -s -X GET http://localhost:8000/v1/auth/keys \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq '.data.keys | length'
```

---

### 1.16 Revoke API Key — Non-existent ID

```bash
curl -s -X DELETE http://localhost:8000/v1/auth/keys/999999 \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Expected:** HTTP 404, reason `API_KEY_NOT_FOUND`

---

### 1.17 Revoke API Key — Invalid ID Format

```bash
curl -s -X DELETE http://localhost:8000/v1/auth/keys/abc \
  -H "Authorization: Bearer $ACCESS_TOKEN" | jq .
```

**Expected:** HTTP 400, reason `INVALID_ID`

---

### 1.18 Revoke API Key — No Auth

```bash
curl -s -X DELETE http://localhost:8000/v1/auth/keys/1 | jq .
```

**Expected:** HTTP 401, reason `MISSING_TOKEN`

---

### 1.19 AuthMiddleware — Malformed Authorization Header

```bash
# No "Bearer" prefix
curl -s -X GET http://localhost:8000/v1/auth/keys \
  -H "Authorization: Token $ACCESS_TOKEN" | jq .
```

**Expected:** HTTP 401, reason `INVALID_AUTH_HEADER`

```bash
# Empty bearer value
curl -s -X GET http://localhost:8000/v1/auth/keys \
  -H "Authorization: Bearer " | jq .
```

**Expected:** HTTP 401, reason `INVALID_TOKEN`

---

### 1.20 AuthMiddleware — Refresh Token as Access Token

```bash
curl -s -X GET http://localhost:8000/v1/auth/keys \
  -H "Authorization: Bearer $REFRESH_TOKEN" | jq .
```

**Expected:** HTTP 401, reason `INVALID_TOKEN_KIND`

---

### 1.21 Health Endpoint Still Works

```bash
curl -s http://localhost:8000/health | jq .
```

**Expected:** HTTP 200, `{"status":"ok"}`

---

## Part 2: Test Case Matrix

### 2.1 POST /v1/auth/login

| # | Case | Input | Expected Status | Expected Reason |
|---|---|---|---|---|
| L-01 | Valid credentials | `email=demo@pulse.dev, password=demo1234` | 200 | — |
| L-02 | Wrong password | `email=demo@pulse.dev, password=wrong` | 401 | `INVALID_CREDENTIALS` |
| L-03 | Non-existent email | `email=nobody@x.com, password=demo1234` | 401 | `INVALID_CREDENTIALS` |
| L-04 | Empty email | `email="", password=demo1234` | 400 | `INVALID_INPUT` |
| L-05 | Empty password | `email=demo@pulse.dev, password=""` | 400 | `INVALID_INPUT` |
| L-06 | Empty body `{}` | — | 400 | `INVALID_INPUT` |
| L-07 | No body (no Content-Type) | — | 400 | `INVALID_BODY` |
| L-08 | Disabled user account | (requires DB update: `UPDATE users SET is_active=false`) | 401 | `ACCOUNT_DISABLED` |

**L-01 additional checks:**
- `data.access_token` is valid JWT with `kind=user`
- `data.refresh_token` is valid JWT with `kind=refresh`
- `data.expires_in = 900`
- Audit log row created with `action=auth.login`

---

### 2.2 POST /v1/auth/refresh

| # | Case | Input | Expected Status | Expected Reason |
|---|---|---|---|---|
| R-01 | Valid refresh token | `refresh_token=<valid-refresh-jwt>` | 200 | — |
| R-02 | Access token used as refresh | `refresh_token=<access-jwt>` | 401 | `INVALID_TOKEN_KIND` |
| R-03 | Expired refresh token | `refresh_token=<expired-jwt>` | 401 | `INVALID_TOKEN` |
| R-04 | Garbage string | `refresh_token=abc123` | 401 | `INVALID_TOKEN` |
| R-05 | Empty field | `refresh_token=""` | 400 | `INVALID_INPUT` |
| R-06 | Empty body `{}` | — | 400 | `INVALID_INPUT` |
| R-07 | Disabled user (refresh) | (DB: `is_active=false`, then use valid refresh token) | 401 | `ACCOUNT_DISABLED` |

**R-01 additional checks:**
- New `access_token` differs from old one
- New `refresh_token` differs from old one
- `expires_in = 900`

---

### 2.3 POST /v1/auth/keys (protected)

| # | Case | Input | Expected Status | Expected Reason |
|---|---|---|---|---|
| K-01 | Create key with name | `name=test-key` + valid Bearer | 200 | — |
| K-02 | Create key with name + expiry | `name=exp-key, expires_at=2026-12-31T...` + valid Bearer | 200 | — |
| K-03 | Missing name | `{}` + valid Bearer | 400 | `INVALID_INPUT` |
| K-04 | No Authorization header | `name=test-key` | 401 | `MISSING_TOKEN` |
| K-05 | Refresh token as Bearer | `name=test-key` + refresh token | 401 | `INVALID_TOKEN_KIND` |
| K-06 | Invalid Bearer token | `name=test-key` + garbage token | 401 | `INVALID_TOKEN` |
| K-07 | No body | valid Bearer, no JSON body | 400 | `INVALID_BODY` |

**K-01 additional checks:**
- `data.raw_key` starts with `pk_` and is 51 chars
- `data.raw_key` is only returned at creation time (never in list)
- `data.id` is a positive integer
- Audit log row created with `action=api_key.create`

---

### 2.4 GET /v1/auth/keys (protected)

| # | Case | Input | Expected Status | Expected Reason |
|---|---|---|---|---|
| KL-01 | List keys with valid token | valid Bearer | 200 | — |
| KL-02 | No Authorization header | — | 401 | `MISSING_TOKEN` |
| KL-03 | Invalid token | garbage Bearer | 401 | `INVALID_TOKEN` |
| KL-04 | Refresh token as Bearer | refresh token | 401 | `INVALID_TOKEN_KIND` |

**KL-01 additional checks:**
- `data.keys` is an array
- Each item has `id`, `name`, `created_at`
- No item contains `raw_key` or `key_hash`
- Newly created keys appear; revoked keys do not

---

### 2.5 DELETE /v1/auth/keys/:id (protected)

| # | Case | Input | Expected Status | Expected Reason |
|---|---|---|---|---|
| KD-01 | Revoke existing key | valid Bearer + existing key ID | 200 | — |
| KD-02 | Revoke non-existent ID | valid Bearer + ID `999999` | 404 | `API_KEY_NOT_FOUND` |
| KD-03 | Invalid ID format | valid Bearer + ID `abc` | 400 | `INVALID_ID` |
| KD-04 | No Authorization header | existing key ID | 401 | `MISSING_TOKEN` |
| KD-05 | Double-revoke same ID | valid Bearer + already-revoked ID | 404 | `API_KEY_NOT_FOUND` |
| KD-06 | Revoke key from another tenant | valid Bearer + key ID belonging to different tenant | 404 | `API_KEY_NOT_FOUND` |

**KD-01 additional checks:**
- Subsequent GET /v1/auth/keys no longer includes the revoked key
- Audit log row created with `action=api_key.revoke`

---

### 2.6 AuthMiddleware

| # | Case | Header | Expected Status | Expected Reason |
|---|---|---|---|---|
| AM-01 | Valid access token | `Authorization: Bearer <access>` | pass-through | — |
| AM-02 | Missing header | (none) | 401 | `MISSING_TOKEN` |
| AM-03 | Empty header value | `Authorization: ` | 401 | `INVALID_AUTH_HEADER` |
| AM-04 | Wrong scheme | `Authorization: Token <jwt>` | 401 | `INVALID_AUTH_HEADER` |
| AM-05 | Expired JWT | `Authorization: Bearer <expired>` | 401 | `INVALID_TOKEN` |
| AM-06 | Malformed JWT | `Authorization: Bearer not.a.jwt` | 401 | `INVALID_TOKEN` |
| AM-07 | Refresh token | `Authorization: Bearer <refresh>` | 401 | `INVALID_TOKEN_KIND` |
| AM-08 | Wrong signing secret | JWT signed with different secret | 401 | `INVALID_TOKEN` |

---

### 2.7 TenantMiddleware

| # | Case | Context | Expected Status | Expected Reason |
|---|---|---|---|---|
| TM-01 | Valid claims with tenant_id | JWT has `attributes.tenant_id` | pass-through | — |
| TM-02 | Missing claims (no AuthMw ran) | No identity in context | 401 | `MISSING_CLAIMS` |
| TM-03 | Claims missing tenant_id | JWT without `attributes.tenant_id` | 401 | `MISSING_TENANT` |
| TM-04 | Non-numeric tenant_id | `attributes.tenant_id = "abc"` | 401 | `INVALID_TENANT` |

---

### 2.8 HMACMiddleware

| # | Case | Header / Body | Expected Status | Expected Reason |
|---|---|---|---|---|
| HM-01 | Valid HMAC signature | Correct `X-Signature: sha256=<hex>` | pass-through | — |
| HM-02 | Missing X-Signature | (none) | 401 | `MISSING_SIGNATURE` |
| HM-03 | Wrong format (no sha256= prefix) | `X-Signature: <hex>` | 401 | `INVALID_SIGNATURE_FORMAT` |
| HM-04 | Invalid hex in signature | `X-Signature: sha256=ZZZZ` | 401 | `INVALID_SIGNATURE_HEX` |
| HM-05 | Wrong HMAC (tampered body) | Valid format but wrong hash | 401 | `INVALID_SIGNATURE` |
| HM-06 | Correct HMAC, body preserved | Check downstream handler receives full body | 200 | — |

**Testing HM-01 manually:**
```bash
BODY='{"service":"test","level":"error","message":"disk full"}'
SECRET="hmac_demo_secret_change_me"
SIG=$(echo -n "$BODY" | openssl dgst -sha256 -hmac "$SECRET" | awk '{print $2}')
echo "X-Signature: sha256=$SIG"
```

---

## Part 3: End-to-End Flow (step-by-step)

Run the full flow in order. Each step depends on the previous one.

```bash
API=http://localhost:8000

# ── Step 1: Login ──────────────────────────────────────────────
echo "=== Step 1: Login ==="
LOGIN=$(curl -s -X POST $API/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"demo@pulse.dev","password":"demo1234"}')
echo "$LOGIN" | jq .

ACCESS=$(echo "$LOGIN" | jq -r '.data.access_token')
REFRESH=$(echo "$LOGIN" | jq -r '.data.refresh_token')
echo "ACCESS=$ACCESS"
echo "REFRESH=$REFRESH"

# ── Step 2: Refresh ────────────────────────────────────────────
echo "=== Step 2: Refresh ==="
REFRESHED=$(curl -s -X POST $API/v1/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH\"}")
echo "$REFRESHED" | jq .

ACCESS=$(echo "$REFRESHED" | jq -r '.data.access_token')
echo "New ACCESS=$ACCESS"

# ── Step 3: Create API Key ─────────────────────────────────────
echo "=== Step 3: Create API Key ==="
CREATED=$(curl -s -X POST $API/v1/auth/keys \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer $ACCESS" \
  -d '{"name":"e2e-test-key"}')
echo "$CREATED" | jq .

KEY_ID=$(echo "$CREATED" | jq -r '.data.id')
RAW_KEY=$(echo "$CREATED" | jq -r '.data.raw_key')
echo "KEY_ID=$KEY_ID  RAW_KEY=$RAW_KEY"

# ── Step 4: List API Keys ──────────────────────────────────────
echo "=== Step 4: List API Keys ==="
curl -s -X GET $API/v1/auth/keys \
  -H "Authorization: Bearer $ACCESS" | jq .

# ── Step 5: Revoke API Key ─────────────────────────────────────
echo "=== Step 5: Revoke API Key ==="
curl -s -X DELETE $API/v1/auth/keys/$KEY_ID \
  -H "Authorization: Bearer $ACCESS" | jq .

# ── Step 6: Verify key is gone ─────────────────────────────────
echo "=== Step 6: Verify key removed from list ==="
curl -s -X GET $API/v1/auth/keys \
  -H "Authorization: Bearer $ACCESS" | jq '.data.keys'

# ── Step 7: Verify 401 without token ──────────────────────────
echo "=== Step 7: Protected endpoint without auth ==="
curl -s -X GET $API/v1/auth/keys | jq .

# ── Step 8: Verify refresh token rejected as access ────────────
echo "=== Step 8: Refresh token as Bearer ==="
curl -s -X GET $API/v1/auth/keys \
  -H "Authorization: Bearer $REFRESH" | jq .
```

**Expected results per step:**

| Step | Expected Status | Key Check |
|---|---|---|
| 1 | 200 | `access_token` and `refresh_token` present |
| 2 | 200 | New token pair returned |
| 3 | 200 | `raw_key` starts with `pk_`, `id` > 0 |
| 4 | 200 | `keys` array includes the new key |
| 5 | 200 | Success message |
| 6 | 200 | `keys` array does NOT include the revoked key |
| 7 | 401 | Reason `MISSING_TOKEN` |
| 8 | 401 | Reason `INVALID_TOKEN_KIND` |

---

## Part 4: Automated Smoke Test

Run the pre-built smoke script:

```bash
./scripts/smoke-auth.sh
```

Override API base if needed:

```bash
API_BASE=http://localhost:8000 ./scripts/smoke-auth.sh
```

The script runs 6 checks and exits non-zero on first failure.

---

## Part 5: Database Verification

After running the e2e flow, verify state directly in Postgres:

```bash
# Connect to DB
PGPASSWORD=postgres psql -h localhost -p 5434 -U postgres -d pulse_db
```

```sql
-- 1. Check audit logs were created
SELECT id, tenant_id, actor_id, action, resource, resource_id, created_at
FROM audit_logs
ORDER BY created_at DESC
LIMIT 10;
-- Expected: rows with action = 'auth.login', 'api_key.create', 'api_key.revoke'

-- 2. Check API key was soft-deleted (not hard-deleted)
SELECT id, tenant_id, name, deleted_at
FROM api_keys
WHERE name = 'e2e-test-key';
-- Expected: deleted_at IS NOT NULL

-- 3. Verify key_hash was never the raw key
SELECT id, key_hash, LENGTH(key_hash) AS hash_len
FROM api_keys
WHERE tenant_id = (SELECT id FROM tenants WHERE slug = 'demo' LIMIT 1);
-- Expected: key_hash is 64-char hex string (SHA-256), never starts with 'pk_'

-- 4. Verify user password is bcrypt hash (not plaintext)
SELECT email, LEFT(password_hash, 7) AS hash_prefix
FROM users
WHERE email = 'demo@pulse.dev';
-- Expected: hash_prefix = '$2a$12$'
```

---

## Part 6: Security Checks

| # | Check | How to Verify | Expected |
|---|---|---|---|
| SEC-01 | Refresh token cannot access protected endpoints | Step 1.20 / Step 8 above | 401 `INVALID_TOKEN_KIND` |
| SEC-02 | Expired tokens are rejected | Wait 15 min or craft an expired JWT | 401 `INVALID_TOKEN` |
| SEC-03 | Raw API key never in list response | Step 1.13 — inspect `data.keys[]` fields | No `raw_key` or `key_hash` field |
| SEC-04 | Password hash never in any response | Inspect all response bodies | No `password_hash` field |
| SEC-05 | key_hash is SHA-256, not plaintext | DB check Part 5, query #3 | 64-char hex, not `pk_...` |
| SEC-06 | Bcrypt used for passwords | DB check Part 5, query #4 | Hash starts with `$2a$` |
| SEC-07 | Tokens use HS256 signing | Decode JWT header at jwt.io | `"alg":"HS256"` |
| SEC-08 | Cross-tenant key revoke blocked | Revoke a key ID from another tenant | 404 `API_KEY_NOT_FOUND` |
| SEC-09 | HMAC body tampering detected | Alter body after signing | 401 `INVALID_SIGNATURE` |
| SEC-10 | JWT signed with wrong secret rejected | Craft JWT with different secret | 401 `INVALID_TOKEN` |
