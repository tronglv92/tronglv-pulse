# Pulse — AI-Native Observability Platform

Pulse ingests logs over HTTP, detects anomalies via rolling z-score, generates AI-powered Root Cause Analysis using Claude Sonnet, and streams live data via WebSocket.

## Stack

Go · Kafka · TimescaleDB (Postgres 15 + pgvector) · Redis · React + Vite

## Quick start

```bash
cp .env.example .env          # fill in secrets
docker compose up -d          # start Postgres, Redis, Kafka
make run                      # API server on :8000 (gRPC :8001)
make ingest                   # Ingest server on :8002
make worker                   # Kafka consumer worker
make cron                     # Scheduled jobs
```

## Docs

- [Architecture](docs/architecture.md)
- [Deployment guide](docs/deployment-guide.md)
- [Task list](docs/task.md)
# tronglv-pulse
