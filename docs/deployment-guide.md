# Pulse — Deployment Guide

**Project:** Pulse — AI-Native Observability Platform
**Last updated:** 2026-05-11

---

## 1. Prerequisites

### Local machine

| Tool | Version | Install |
|---|---|---|
| Go | 1.24+ | `brew install go` |
| Docker | 24+ | [docker.com](https://docker.com) |
| Docker Compose | v2 (bundled with Docker Desktop) | — |
| Make | any | `brew install make` |
| protoc | 3.21+ | `brew install protobuf` |
| protoc-gen-go | latest | `go install google.golang.org/protobuf/cmd/protoc-gen-go@latest` |
| protoc-gen-go-grpc | latest | `go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest` |
| protoc-gen-grpc-gateway | latest | `go install github.com/grpc-ecosystem/grpc-gateway/v2/protoc-gen-grpc-gateway@latest` |
| goctl | latest | `go install github.com/zeromicro/go-zero/tools/goctl@latest` |
| pmctl | latest | Install from internal tooling (`go install ...`) |
| k6 | latest | `brew install k6` (load tests only) |
| mockery | v2 | `go install github.com/vektra/mockery/v2@latest` |

### Production VPS (Hetzner CCX13)

- 4 vCPU, 16 GB RAM, 80 GB SSD
- Ubuntu 24.04 LTS
- Docker 24+ + Docker Compose v2
- Domain pointing to VPS IP (`demo.pulse.dev`)

---

## 2. Local Development Setup

```bash
# 1. Clone
git clone https://github.com/<your-username>/pulse.git && cd pulse

# 2. Copy env template
cp .env.example .env
# Edit .env: add ANTHROPIC_API_KEY, OPENAI_API_KEY (embedding), SLACK_WEBHOOK_URL

# 3. Start infrastructure (Postgres + TimescaleDB + pgvector, Redis, Redpanda, pgAdmin)
make up
# or: docker compose up -d

# 4. Run DB migrations
make migrate-up

# 5. Seed demo data (tenant, user, API key)
make seed

# 6. Start all services (3 separate terminals, or use air for live-reload)
make run-api        # HTTP :8000 + gRPC :8001
make run-ingest     # HTTP :8002
make run-worker     # Kafka consumers
# Alternatively: air (reads .air.toml for live-reload)

# 7. Start the frontend
cd web && npm install && npm run dev
# → http://localhost:3000
```

### Quick health check

```bash
curl http://localhost:8000/health   # api service
curl http://localhost:8002/health   # ingest service
```

---

## 3. Environment Variables Reference

All variables live in `.env` (mode `0600`). They are substituted into `etc/*.yaml` using go-zero's `conf.MustLoad`.

| Variable | Default | Required | Description |
|---|---|---|---|
| `SERVER_ENV` | `local` | yes | Environment label (`local` / `staging` / `production`) |
| `SERVICE_API_HTTP_PORT` | `8000` | yes | api service HTTP port |
| `SERVICE_API_GRPC_PORT` | `8001` | yes | api service gRPC port |
| `SERVICE_INGEST_HTTP_PORT` | `8002` | yes | ingest service HTTP port |
| `DB_DRIVER` | `postgres` | yes | Database driver |
| `DB_HOST` | `localhost` | yes | Postgres host |
| `DB_PORT` | `5432` | yes | Postgres port |
| `DB_NAME` | `pulse_db` | yes | Database name |
| `DB_USERNAME` | `postgres` | yes | Database user |
| `DB_PASSWORD` | `postgres` | yes | Database password |
| `DB_SCHEMA_NAME` | `public` | yes | Postgres schema |
| `REDIS_HOST` | `localhost:6379` | yes | Redis address |
| `REDIS_PASSWORD` | `` | no | Redis password |
| `REDIS_DB` | `0` | yes | Redis DB index |
| `KAFKA_BROKER_HOST` | `localhost:9092` | yes | Redpanda/Kafka broker |
| `JWT_SECRET` | — | yes | HS256 signing secret (min 32 chars) |
| `INGEST_HMAC_SECRET` | — | yes | HMAC SHA-256 shared secret for `/v1/logs` |
| `ANTHROPIC_API_KEY` | — | yes | Claude API key |
| `ANTHROPIC_MODEL` | `claude-sonnet-4-6` | yes | LLM model ID |
| `OPENAI_API_KEY` | — | yes | OpenAI key for `text-embedding-3-small` |
| `SLACK_WEBHOOK_URL` | — | no | Slack incoming webhook URL |
| `LLM_DAILY_BUDGET_USD` | `5.0` | yes | Hard daily LLM spend cap |
| `TELEMETRY_ENDPOINT` | `` | no | OpenTelemetry collector endpoint |

---

## 4. Database Initialization

```bash
# Install golang-migrate CLI
go install -tags 'postgres' github.com/golang-migrate/migrate/v4/cmd/migrate@latest

# Run all migrations up
make migrate-up
# Equivalent to:
migrate -path db/migrations -database "postgres://postgres:postgres@localhost:5432/pulse_db?sslmode=disable" up

# Roll back one migration
make migrate-down

# Check current version
migrate -path db/migrations -database "..." version
```

### Verify extensions installed

```sql
-- Connect: psql -U postgres pulse_db
SELECT extname FROM pg_extension;
-- Expected: timescaledb, vector, pg_trgm, citext, plpgsql
```

---

## 5. Running the Demo

The full demo follows the 5-step story from the BA document. All steps use Makefile targets.

```bash
# Step 1: Full reset + seed + start traffic simulation
make demo
# (Equivalent to: make migrate-down && make migrate-up && make seed && make simulate-traffic)

# Step 2: Trigger an incident (high error rate in checkout service)
make break-checkout
# Runs sim with --service=checkout --error-rate=0.5 for 30s

# Step 3: Wait ~60 seconds → dashboard turns red

# Step 4: Click incident card in browser → read AI RCA → click Find Similar

# Step 5: Restore normal traffic
make fix-checkout
# Runs sim with --service=checkout --error-rate=0.01

# Step 6: Resolve the incident in the UI (requires category + note)
```

### Demo script flags

```bash
# Custom traffic simulation
go run cmd/sim/main.go \
  --target http://localhost:8002 \
  --api-key pk_demo_xxx \
  --hmac-secret <secret> \
  --service checkout \
  --rps 200 \
  --error-rate 0.5 \
  --duration 60s
```

---

## 6. Production Deployment (Hetzner VPS)

### 6.1 Provision the VPS

1. Create Hetzner CCX13 in Falkenstein (eu-central) with Ubuntu 24.04.
2. Add your SSH public key during creation.
3. Configure DNS: `demo.pulse.dev` A record → VPS public IP.

### 6.2 Initial VPS setup

```bash
ssh root@<vps-ip>

# Install Docker
curl -fsSL https://get.docker.com | sh
usermod -aG docker ubuntu

# UFW firewall (only 22/80/443)
ufw allow OpenSSH
ufw allow 80/tcp
ufw allow 443/tcp
ufw enable

# Install Docker Compose plugin
apt-get install -y docker-compose-plugin

# Verify
docker compose version
```

### 6.3 Deploy the application

```bash
# On VPS
git clone https://github.com/<your-username>/pulse.git /opt/pulse
cd /opt/pulse

# Create production .env (mode 0600, never commit this file)
cp .env.example .env
chmod 0600 .env
nano .env  # Fill in all production values

# Start all services
docker compose -f deploy/docker-compose.yml up -d

# Verify health
docker compose -f deploy/docker-compose.yml ps
curl https://demo.pulse.dev/health
```

### 6.4 Updating the deployment

```bash
cd /opt/pulse
git pull
docker compose -f deploy/docker-compose.yml pull
docker compose -f deploy/docker-compose.yml up -d --remove-orphans
make migrate-up  # run from VPS if schema changed
```

---

## 7. Caddy TLS Setup

`deploy/Caddyfile`:

```caddyfile
demo.pulse.dev {
    # Ingest service
    handle /v1/logs* {
        reverse_proxy ingest:8002
    }

    # API service (REST + WebSocket)
    handle /v1/* {
        reverse_proxy api:8000
    }

    # Frontend
    handle {
        reverse_proxy web:3000
    }

    # Security headers
    header {
        Strict-Transport-Security "max-age=31536000; includeSubDomains"
        X-Content-Type-Options "nosniff"
        X-Frame-Options "DENY"
        Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline'; style-src 'self' 'unsafe-inline'; connect-src 'self' wss://demo.pulse.dev"
    }
}
```

Caddy handles ACME Let's Encrypt certificate automatically. No manual cert management required.

---

## 8. Monitoring

### Prometheus

Prometheus scrapes all three services at `/metrics` every 15s. Config at `deploy/prometheus.yml`:

```yaml
global:
  scrape_interval: 15s

scrape_configs:
  - job_name: pulse-api
    static_configs:
      - targets: ['api:8000']
  - job_name: pulse-ingest
    static_configs:
      - targets: ['ingest:8002']
  - job_name: pulse-worker
    static_configs:
      - targets: ['worker:8003']
  - job_name: redpanda
    static_configs:
      - targets: ['redpanda:9644']  # Redpanda Prometheus port
```

### Grafana

Access Grafana at `http://localhost:3001` (local) or behind a Caddy sub-path in production.

Two pre-built dashboards shipped in `deploy/grafana/dashboards/`:

| Dashboard | Key panels |
|---|---|
| **Pulse Internals** | Ingest RPS, p99 ingest latency, Kafka consumer lag, DB active connections, Redis memory, service error rate |
| **Demo Tenant** | Log ingestion rate, active incidents count, RCA cache hit rate, LLM daily spend, anomaly detection cycle time |

Import dashboards:
```bash
# Grafana auto-provisions from deploy/grafana/provisioning/
# No manual import needed if using docker-compose.yml
```

---

## 9. CI/CD Pipeline

GitHub Actions workflow at `.github/workflows/ci.yml`:

```
On: push (all branches) + pull_request (→ main)

Jobs:
  test:
    1. Set up Go 1.24
    2. Cache Go modules
    3. golangci-lint run (default config)
    4. go test -race ./...
    5. go test -bench=. -benchmem ./cmd/ingest/... (ingest only)

  build:
    1. docker build -f Dockerfile         --tag pulse-api:$SHA
    2. docker build -f DockerfileIngest   --tag pulse-ingest:$SHA
    3. docker build -f DockerfileWorker   --tag pulse-worker:$SHA
    4. On tag push: push all images to ghcr.io

Deploy: manual SSH + docker compose pull && up -d
```

**CD is intentionally manual** — SSH to VPS and run `git pull && docker compose up -d`. A watchtower or ArgoCD setup is documented as future work.

---

## 10. Load Testing

```bash
# Prerequisite: services running at https://demo.pulse.dev (or localhost:8002)

# Steady-state: 10k logs/sec for 5 minutes
k6 run scripts/load-test/ingest-steady.js \
  -e TARGET=https://demo.pulse.dev \
  -e API_KEY=pk_demo_xxx \
  -e HMAC_SECRET=<secret>

# Burst: 20k logs/sec for 30 seconds
k6 run scripts/load-test/ingest-burst.js \
  -e TARGET=https://demo.pulse.dev

# WebSocket concurrency: 100 simultaneous WS connections
k6 run scripts/load-test/ws-concurrency.js \
  -e TARGET=wss://demo.pulse.dev
```

### Recording benchmark results

After each k6 run, save the summary to `docs/benchmarks/`:

```bash
k6 run scripts/load-test/ingest-steady.js \
  --out json=docs/benchmarks/$(date +%Y-%m-%d)-steady.json \
  --summary-export=docs/benchmarks/$(date +%Y-%m-%d)-summary.json
```

Target numbers to verify (add to README):
- Steady 10k logs/sec × 5 min: **p99 < 80ms** ✓
- Burst 20k logs/sec × 30s: **p99 < 200ms** ✓
- 100 concurrent WS connections: **no message loss** ✓

---

## 11. Troubleshooting

### Redpanda not ready after `docker compose up`

```bash
# Check Redpanda health
docker compose logs redpanda | tail -30
# Wait ~15s for KRaft leader election before starting worker/ingest

# Manually create topics if auto-create is off
docker compose exec redpanda rpk topic create logs.raw logs.enriched anomalies.detected rca.completed \
  --partitions 4 --replicas 1
```

### Postgres extensions missing

```bash
docker compose exec postgres psql -U postgres pulse_db -c "\dx"
# If timescaledb or vector are missing, re-run migrations:
make migrate-down && make migrate-up
# If that fails, check the image: must be timescale/timescaledb-ha:pg16-all
```

### LLM daily budget hit (`llm:disabled=true` in Redis)

```bash
# Check the flag
docker compose exec redis redis-cli GET llm:disabled

# Reset manually (resets at midnight automatically)
docker compose exec redis redis-cli DEL llm:disabled

# Check today's spend
psql -c "SELECT sum(cost_usd) FROM llm_calls WHERE created_at >= CURRENT_DATE"
```

### WebSocket connections dropping

- Check JWT expiry — access tokens are 15 minutes; React must refresh before expiry.
- Check Redpanda consumer lag — if `logs.enriched` lag > 10k, the bridge may be overwhelmed; reduce ingest rate.
- Check worker logs for panic recovery messages: `docker compose logs worker`.

### Migration version mismatch

```bash
# Check current version in DB
migrate -path db/migrations -database "postgres://..." version

# Force a version if dirty state
migrate -path db/migrations -database "postgres://..." force <version>
```

---

## 12. Future Work (documented here, not built)

These are referenced in the README "What I'd build next" section:

| Item | Rough effort |
|---|---|
| Kubernetes deployment with Helm chart (`deploy/helm/`) | ~15h |
| ArgoCD GitOps pipeline | ~8h |
| Managed Postgres (Neon / Supatera) instead of self-hosted | ~3h |
| Debezium CDC replacing outbox polling (sub-100ms event latency) | ~10h |
| Temporal for saga orchestration (instead of in-process executor) | ~15h |
| ClickHouse for logs at 1M+ logs/sec scale | ~20h |
| Multi-tenant RBAC with full RLS | ~20h |
| OpenTelemetry distributed tracing (Jaeger) | ~10h |
| Natural-language → SQL log search | ~25h |

---

**End of Deployment Guide**
