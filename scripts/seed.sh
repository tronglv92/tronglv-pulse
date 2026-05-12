#!/usr/bin/env bash
# seed.sh — insert demo tenant, user (bcrypt pw), API key + HMAC secret
set -euo pipefail

: "${DB_HOST:=localhost}"
: "${DB_PORT:=5433}"
: "${DB_NAME:=pulse_db}"
: "${DB_USER:=postgres}"
: "${DB_PASSWORD:=postgres}"

PGPASSWORD="$DB_PASSWORD" psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<SQL
-- TODO TASK-003: insert demo tenant, user, api_key rows
SELECT 'seed placeholder';
SQL

echo "Seed complete."
