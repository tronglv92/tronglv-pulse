#!/usr/bin/env bash
# seed.sh — insert demo tenant, user (bcrypt pw), and API key (idempotent)
#
# Demo credentials:
#   Tenant slug : demo
#   HMAC secret : hmac_demo_secret_change_me
#   User email  : demo@pulse.dev
#   Password    : demo1234  (bcrypt cost-12)
#   API key     : pk_demo_0000000000000000  (SHA-256 stored in key_hash)
#
# Usage:
#   ./scripts/seed.sh
#   DB_HOST=prod-host DB_PASSWORD=secret ./scripts/seed.sh
set -euo pipefail

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5433}"
: "${DB_NAME:=pulse_db}"
: "${DB_USER:=postgres}"
: "${DB_PASSWORD:=postgres}"

# Pre-computed values (avoid bcrypt/sha256 shell dependencies):
#   bcrypt("demo1234", cost=12)
DEMO_PASSWORD_HASH='$2a$12$/uAjrxkQzjW2BMo.C7w4nOObNbDHJ/Rr/k5Tf.7NfQfFfFc844yPO'
#   sha256("pk_demo_0000000000000000")
DEMO_KEY_HASH='6b4c8b0a7b2b8454cc4a737bf4c8b2d84f9efd791989dcca6f820aff274dcc35'

PGPASSWORD="$DB_PASSWORD" psql \
  -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" \
  -v ON_ERROR_STOP=1 \
  <<SQL

-- ── 1. Demo tenant ───────────────────────────────────────────────────────────
INSERT INTO tenants (name, slug, api_secret, is_active, created_at, updated_at)
VALUES (
  'Demo Tenant',
  'demo',
  'hmac_demo_secret_change_me',
  true,
  NOW(), NOW()
)
ON CONFLICT (slug) WHERE deleted_at IS NULL
DO UPDATE SET
  api_secret = EXCLUDED.api_secret,
  name       = EXCLUDED.name,
  updated_at = NOW();

-- ── 2. Demo user (bcrypt cost-12 of "demo1234") ──────────────────────────────
INSERT INTO users (tenant_id, email, password_hash, name, is_active, created_at, updated_at)
SELECT
  t.id,
  'demo@pulse.dev',
  '$DEMO_PASSWORD_HASH',
  'Demo User',
  true,
  NOW(), NOW()
FROM tenants t
WHERE t.slug = 'demo' AND t.deleted_at IS NULL
ON CONFLICT (email)
DO UPDATE SET
  password_hash = EXCLUDED.password_hash,
  name          = EXCLUDED.name,
  updated_at    = NOW();

-- ── 3. Demo API key (sha256 of "pk_demo_0000000000000000") ──────────────────
INSERT INTO api_keys (tenant_id, key_hash, name, created_at, updated_at)
SELECT
  t.id,
  '$DEMO_KEY_HASH',
  'Demo Key',
  NOW(), NOW()
FROM tenants t
WHERE t.slug = 'demo' AND t.deleted_at IS NULL
ON CONFLICT (key_hash) WHERE deleted_at IS NULL
DO UPDATE SET
  name       = EXCLUDED.name,
  updated_at = NOW();

-- ── Summary ──────────────────────────────────────────────────────────────────
SELECT
  t.id   AS tenant_id,
  t.slug,
  t.api_secret                        AS hmac_secret,
  u.email,
  'demo1234'                          AS password_plaintext,
  'pk_demo_0000000000000000'          AS api_key_plaintext,
  k.key_hash
FROM tenants t
JOIN users   u ON u.tenant_id = t.id AND u.deleted_at IS NULL
JOIN api_keys k ON k.tenant_id = t.id AND k.deleted_at IS NULL
WHERE t.slug = 'demo' AND t.deleted_at IS NULL;

SQL

echo "Seed complete."
