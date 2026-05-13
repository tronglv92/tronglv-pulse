# Pulse — Sprint Task List

**Project:** Pulse — AI-Native Observability Platform (Portfolio Edition)
**Total budget:** ~150 hours over 12 weeks (~2.5 hrs/day, 5 days/week)
**Last updated:** 2026-05-13 (done: TASK-001–008)

---

## How to Read This File

| Field | Meaning |
|---|---|
| **ID** | Unique task identifier (`TASK-NNN`) |
| **Est** | Estimated hours |
| **Pri** | `must` = non-negotiable for demo, `should` = cut if over budget |
| **Deps** | Tasks that must complete first |
| **Status** | `todo` / `in-progress` / `done` |

### Execution philosophy

1. **Shared infrastructure first** — config, registry, Kafka, middleware, external clients. No domain code until these compile.
2. **One reference domain end-to-end (`auth`)** — complete vertical slice: middleware → handler → service → repository. No Kafka, no AI. Clean pattern for all other domains to follow.
3. **Remaining domains in ascending complexity** — search → ingestion → incident → notification → stream → rca.
4. When implementing a domain, read `internal/auth/` first as the reference implementation.

The **non-negotiable core chain**: `TASK-001 → TASK-004 → TASK-005 → TASK-006 → TASK-009 → TASK-011 → TASK-023 → TASK-029 → TASK-035 → TASK-038 → TASK-043 → TASK-049 → TASK-053 → TASK-057 → TASK-066`

---

## Phase 0: Foundation (done)

All TASK-001 through TASK-007 are complete.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-001 | Init Go monorepo: `go.mod` (module `pulse`), `cmd/` binaries, `go.work`, `pmctl.gen.yaml`, `CLAUDE.md`, `.air.toml`, `.mcp.json`, `.env.example`, `.gitignore` | 2h | must | — | done |
| TASK-002 | `docker-compose.yml`: TimescaleDB + pgvector (`pg16-all`), Redis 7, Redpanda, pgAdmin; bind ports 5433/6379/9092/5050 on `pulse` bridge network | 2h | must | TASK-001 | done |
| TASK-003 | DB migrations 0001–0009 (`db/migrations/`): extensions (timescaledb, vector, pg_trgm, citext), log_entries hypertable, `logs_metrics_5min` continuous aggregate, incidents, error_embeddings, llm_calls, rca_cache, api_keys, outbox_events, saga_instances, audit_logs; both `.up.sql` and `.down.sql` | 5h | must | TASK-002 | done |
| TASK-004 | GORM entities (`internal/types/entity/`): `LogEntry`, `AnomalyEvent`, `Incident`, `ErrorEmbedding`, `LLMCall`, `RCACache`, `Tenant`, `User`, `APIKey`, `OutboxEvent`, `SagaInstance`, `AuditLog`; register in `migrate.go`; constants in `internal/types/define/constant/`; enums in `internal/types/define/enum/` | 3.5h | must | TASK-003 | done |
| TASK-005 | Proto contracts (`api/`): `ingest.proto`, `incident.proto`, `search.proto`, `stream.proto`, `auth.proto`, `common.proto`; vendor shared protos into `protos/`; run `make grpc && make validate` | 3h | must | TASK-001 | done |
| TASK-006 | Port interfaces (`internal/contract/`): `repository.go` (9 repo interfaces), `tx.go` (TxManager), `outbox.go` (OutboxAppender + OutboxRepository), `publisher.go` (KafkaPublisher), `llm.go` (LLMClient), `cache.go` (Cache), `notifier.go` (Notifier), `stats.go` (StatsProvider), `saga.go` (Step + Saga + SagaRepository) | 2h | must | TASK-004 | done |
| TASK-007 | Restructure `internal/` into domain modules: create `ingestion/`, `incident/`, `rca/`, `search/`, `stream/`, `auth/`, `notification/`; each owns handler + service + repository + consumer + mapper + dto + event; delete `service/`, `server/`, `types/dto/`, `types/event/` | 2h | must | TASK-006 | done |

---

## Phase 1: Shared Infrastructure (~20 hrs)

**Goal:** All three binaries compile and return `/health`. Kafka, outbox, external clients, stats helpers all in place. No domain-specific code yet.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-008 | Config structs (`internal/config/config.go`): one `Config` per binary; `etc/api.yaml`, `etc/ingest.yaml`, `etc/worker.yaml`, `etc/cron.yaml` with `${ENV_VAR}` substitution; complete `.env.example` with all variables and defaults | 1.5h | must | TASK-005 | done |
| TASK-009 | `GET /health` on api (8000) and ingest (8002): no auth, no DB call; `cmd/api/main.go` and `cmd/ingest/main.go` wire go-zero ServiceGroup; verify `curl localhost:{8000,8002}/health` → 200 | 1h | must | TASK-008 | todo |
| TASK-010 | `internal/registry/` DI composition root: `base_context.go` (BaseContext: GetConfig, GetDownloader), `repository_context.go` (all repo getters — stubs for now), `service_context.go` (HTTP/gRPC wiring), `consumer_context.go` (Kafka consumer wiring), `cron_context.go` (cron job wiring), `security_authentication.go`, `security_authorization.go`, `security_http.go`, `security_context.go` | 3h | must | TASK-008 | todo |
| TASK-011 | Makefile: `make up/down`, `make run-api/ingest/worker/cron`, `make air`, `make test/coverage`, `make grpc/validate/mocks`, `make migrate-up/down`, `make seed`, `make demo`; Dockerfiles for each binary | 1.5h | must | TASK-001 | todo |
| TASK-012 | GitHub Actions CI (`.github/workflows/ci.yml`): `golangci-lint`, `go test -race ./...`, `docker build` all images; push to `ghcr.io` on tag; CI badge in README | 2h | must | TASK-001 | todo |
| TASK-013 | Seed script (`scripts/seed.sh`): insert demo tenant, demo user (bcrypt cost-12), demo API key (`pk_demo_xxx`) with known HMAC secret; idempotent (upsert) | 1h | should | TASK-003 | todo |
| TASK-014 | Cross-cutting middleware (`internal/middleware/`): `logging_middleware.go` (structured request + response log), `recovery_middleware.go` (panic → 500), `rate_limit_middleware.go` (Redis token-bucket 50k/sec per tenant) | 1.5h | must | TASK-010 | todo |
| TASK-015 | Kafka producer (`internal/kafka/producer.go`): wraps `segmentio/kafka-go`; idempotent batching, `acks=all`; `internal/kafka/topics.go` (5 topic constants: `pulse.logs.raw`, `pulse.anomalies`, `pulse.rca.request`, `pulse.rca.result`, `pulse.outbox`) | 2h | must | TASK-010 | todo |
| TASK-016 | Kafka consumer scaffold (`internal/kafka/consumer.go`): wraps `kq.MustNewQueue`; `ConsumeHandler` interface; `internal/consumer/consumer.go` supervisor — one goroutine per consumer with panic-recovery restart | 2h | must | TASK-010 | todo |
| TASK-017 | Outbox publisher (`internal/outbox/publisher.go`): goroutine in `cmd/worker`; `FOR UPDATE SKIP LOCKED` fetches ≤100 unpublished rows every 500ms; publishes via `KafkaPublisher`; exponential backoff on failure; `MarkPublished` on success; `internal/outbox/repository.go` implements `contract.OutboxRepository`; dedup via `internal/outbox/dedup.go` (Redis key `evt:{id}`, 1h TTL) | 3h | must | TASK-015 TASK-016 | todo |
| TASK-018 | External clients: `external/anthropic/client.go` (implements `contract.LLMClient.Generate`, 20s timeout, retry once on parse failure); `external/openai/client.go` (implements `LLMClient.Embed`, 1536 dims); `external/slack/client.go` (implements `contract.Notifier.SendAlert`, retry × 5 with backoff); `internal/llm/fake.go` (deterministic output by input hash, for tests) | 3h | must | TASK-006 | todo |
| TASK-019 | Stats helpers (`helper/stats/`): `zscore.go` (rolling z-score with configurable baseline window, sparse-data fallback returns 0); `ewma.go`; unit tests for both | 2h | must | — | todo |

**Phase 1 exit criteria:** `make up` starts three services; `/health` returns 200 on api and ingest; `make test` passes; `go build ./...` on `internal/` passes clean.

---

## Phase 2: Reference Domain — `auth` (~8 hrs)

**Goal:** Complete `internal/auth/` end-to-end. This is the reference implementation — every other domain follows the same layer pattern. Read this domain before implementing any other.

**Pattern established here:**
- `auth/X_repo.go` implements a `contract.XRepo` interface (GORM, db.WithContext)
- `auth/service.go` depends on repo interfaces, no concrete types
- `auth/handler.go` parses proto request → validates → calls service → returns proto response
- `auth/X_middleware.go` for domain-specific middleware
- Registry wiring: repo → service_context → grpc_handler + http_handler

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-020 | `internal/auth/` repositories: `tenant_repo.go` (TenantRepo impl: Save, FindByID, FindBySlug, Update); `api_key_repo.go` (APIKeyRepo impl: Save, FindByHash, ListByTenant, Delete, TouchLastUsed); `audit_repo.go` (AuditLogRepo impl: Append, FindByResource) — all GORM with `db.WithContext(ctx)` | 2h | must | TASK-010 | todo |
| TASK-021 | `internal/auth/service.go` (AuthService): bcrypt compare via `helper/utils/authenticator/`; issue 15-min access + 30-day refresh JWT; rotation on refresh; create/revoke API key (HMAC key hash stored, raw key returned once); TouchLastUsed on every authenticated request; AuditLog append on sensitive operations | 2h | must | TASK-020 | todo |
| TASK-022 | `internal/auth/` middleware: `hmac_middleware.go` (verify `X-Signature: sha256=<hmac>` against `INGEST_HMAC_SECRET`; 401 on fail); `auth_middleware.go` (JWT Bearer verify; 401 on fail); `tenant_middleware.go` (extract tenant_id from JWT claims → context via `helper/utils/identity/`) | 1.5h | must | TASK-021 | todo |
| TASK-023 | `internal/auth/handler.go`: `POST /v1/auth/login`, `POST /v1/auth/refresh`, `POST /v1/auth/keys` (create), `DELETE /v1/auth/keys/{id}` (revoke), `GET /v1/auth/keys` (list); wire into `internal/handler/grpc_handler.go` and `internal/handler/http_handler.go`; register repos + service in `internal/registry/repository_context.go` and `service_context.go`; run `pmctl gen service` | 1.5h | must | TASK-022 TASK-010 | todo |
| TASK-024 | End-to-end smoke test for auth domain: seed tenant + user → `POST /v1/auth/login` → verify JWT → `POST /v1/auth/keys` → verify HMAC signature with new key → `GET /v1/auth/keys` returns list; script in `scripts/smoke-auth.sh` | 1h | must | TASK-023 TASK-013 | todo |

**Phase 2 exit criteria:** `bash scripts/smoke-auth.sh` passes end-to-end without manual steps. Auth domain is the template — read `internal/auth/` before writing any other domain.

---

## Phase 3: Search Domain (~4 hrs)

**Goal:** Read-only query API for log entries. Simplest domain after auth — no Kafka, no AI. Establishes the read-path pattern.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-025 | `internal/search/` repository + service: search service uses `LogRepo` (from ingestion domain) directly via interface; `search/service.go` implements full-text via `pg_trgm` GIN index, date-range filter, pagination | 1.5h | must | TASK-010 | todo |
| TASK-026 | `internal/search/handler.go`: `GET /v1/logs/search?q=&service=&level=&from=&to=&page=&limit=`; `GET /v1/logs/{id}`; apply JWT + tenant middleware; response uses `internal/search/dto.go` shapes mapped via `internal/search/mapper.go` | 1.5h | must | TASK-025 TASK-022 | todo |
| TASK-027 | Wire search into registry; run `pmctl gen service`; smoke test: seed 50 log rows → search by service name → verify pagination works | 1h | must | TASK-026 TASK-010 | todo |

**Phase 3 exit criteria:** `GET /v1/logs/search?q=error` returns paginated results with correct tenant scoping.

---

## Phase 4: Ingestion Domain (~10 hrs)

**Goal:** `POST /v1/logs` → fingerprint → Postgres → Kafka → enricher. Visible by querying the DB.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-028 | `internal/ingestion/repository.go` (LogRepo impl): batch insert to `log_entries` hypertable via GORM; `CountByWindow` for z-score windowed counts; `db.WithContext(ctx)` on every call | 1.5h | must | TASK-010 | todo |
| TASK-029 | `internal/ingestion/service.go` (IngestService): validate batch (≤500 entries, ≤2MB), compute SHA-256 fingerprint per entry (service + level + top error message), call `KafkaPublisher.Publish(pulse.logs.raw)` via outbox pattern; `internal/ingestion/dto.go` (LogBatch, LogEntry shapes); run `pmctl gen service` | 2h | must | TASK-028 TASK-015 | todo |
| TASK-030 | `internal/ingestion/enricher.go` (EnricherService): parse raw log, insert to log_entries hypertable; call `LogRepo.CountByWindow` for z-score window data, compute z-score via `helper/stats/zscore.go`; if z-score > threshold → publish `AnomalyDetectedEvent` to `pulse.anomalies` via outbox; `internal/ingestion/event.go` (RawLogEvent shape) | 3h | must | TASK-028 TASK-019 TASK-017 | todo |
| TASK-031 | `internal/ingestion/consumer.go`: subscribes to `pulse.logs.raw`; calls EnricherService; idempotency via `outbox/dedup.go`; register in `internal/consumer/consumer.go` supervisor and `internal/registry/consumer_context.go` | 1.5h | must | TASK-030 TASK-016 | todo |
| TASK-032 | `internal/ingestion/handler.go`: `POST /v1/logs` (batch ingest); `internal/ingestion/server.go` (go-zero HTTP server builder for cmd/ingest on port 8002); attach HMAC middleware + tenant middleware + rate-limit middleware from `internal/auth/` and `internal/middleware/` | 1.5h | must | TASK-029 TASK-022 TASK-014 | todo |
| TASK-033 | `cmd/sim/main.go` synthetic traffic generator: `POST /v1/logs` with valid HMAC; flags `--target --api-key --hmac-secret --service --rps --error-rate --duration`; `scripts/simulate-traffic.sh` runs at 100 rps for 60s | 2h | must | TASK-032 | todo |

**Phase 4 exit criteria:** `bash scripts/simulate-traffic.sh` runs 60s → `SELECT count(*) FROM log_entries` > 5000 rows.

---

## Phase 5: Incident Domain (~10 hrs)

**Goal:** Anomaly detected → incident created in DB → ack/resolve lifecycle works.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-034 | `internal/incident/repository.go` (IncidentRepo impl): Create, FindByID, FindByFingerprint, Update, List with filter (status, service, tenant, page/limit) | 1.5h | must | TASK-010 | todo |
| TASK-035 | `internal/incident/anomaly.go` (AnomalyService): validate z-score threshold (configurable, default 3.0); 10-min cooldown per `(tenant, service, fingerprint)` in Redis; insert `Incident` (status=OPEN, z_score, fingerprint, severity) within `TxManager.RunInTx` + `OutboxAppender.Append` in same transaction; `internal/incident/event.go` (AnomalyDetectedEvent shape) | 2.5h | must | TASK-034 TASK-017 TASK-010 | todo |
| TASK-036 | `internal/incident/service.go` (IncidentService): List (filter by status/service), GetByID, Acknowledge (status OPEN → ACK, set acked_at), Resolve (status ACK → RESOLVED, validate note + category, compute MTTR, write via RunInTx + outbox); run `pmctl gen service` | 2h | must | TASK-034 TASK-017 | todo |
| TASK-037 | `internal/incident/handler.go`: `GET /v1/incidents`, `GET /v1/incidents/{id}`, `POST /v1/incidents/{id}/ack`, `POST /v1/incidents/{id}/resolve`; `internal/incident/server.go`; attach JWT + tenant middleware; response uses `internal/incident/dto.go` + `internal/incident/mapper.go` | 1.5h | must | TASK-036 TASK-022 | todo |
| TASK-038 | Cron: `internal/cron/command.go` (CLI command registry keyed by job name); `internal/cron/anomaly_cronjob.go` (ticker 10s, query `logs_metrics_5min` continuous aggregate via GORM, call AnomalyService for each `(tenant, service)` group); add to cron registry and `cmd/cron/main.go` | 2h | must | TASK-035 TASK-010 | todo |
| TASK-039 | `scripts/break-checkout.sh` (sim at `--error-rate=0.5 --service=checkout --duration=30s`) + `scripts/fix-checkout.sh` (reverts to `--error-rate=0.01`); wire incident into registry | 0.5h | must | TASK-033 TASK-038 | todo |

**Phase 5 exit criteria:** `bash scripts/break-checkout.sh` → wait 60s → incident row in DB with status=OPEN and z-score > 3; `POST /v1/incidents/{id}/ack` → status changes to ACK.

---

## Phase 6: Notification Domain (~3 hrs)

**Goal:** RCA result triggers Slack alert. Consumes from `pulse.rca.result`.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-040 | `internal/notification/service.go` (NotificationService): format `AlertPayload` → Slack Block Kit message (service, z-score, RCA summary, dashboard link); call `contract.Notifier.SendAlert`; append `AuditLog` record | 1h | should | TASK-018 TASK-010 | todo |
| TASK-041 | `internal/notification/consumer.go`: subscribes to `pulse.rca.result`; deserializes `RCACompletedEvent`; calls NotificationService; idempotency via dedup; register in consumer supervisor + registry | 1h | should | TASK-040 TASK-016 | todo |
| TASK-042 | Wire notification into registry; smoke test: manually publish a `RCACompletedEvent` to `pulse.rca.result` → verify Slack message arrives | 1h | should | TASK-041 | todo |

**Phase 6 exit criteria:** Slack message arrives within 5s of a manually published `RCACompletedEvent`.

---

## Phase 7: Stream Domain (~5 hrs)

**Goal:** Browser receives live log + incident events over WebSocket without polling.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-043 | `internal/stream/hub.go` (WebSocket hub): in-process pub/sub; typed channels with service filter; fan-out to matching connections; sample beyond 1000 msgs/sec per client; thread-safe with `sync.RWMutex` | 3h | must | TASK-010 | todo |
| TASK-044 | `internal/stream/handler.go`: `GET /v1/stream?token=<jwt>`; JWT validation; subscribe message `{action, channel, filter}`; write `{type, data}` frames; heartbeat ping every 30s; auto-close on auth expiry; Kafka → WS bridge: consumer subscribes to `pulse.logs.raw` + `pulse.anomalies` + `pulse.rca.result` → pushes to hub | 2h | must | TASK-043 TASK-022 | todo |

**Phase 7 exit criteria:** `wscat -c ws://localhost:8000/v1/stream?token=<jwt>` receives live events within 3s of sim traffic.

---

## Phase 8: RCA Domain (~16 hrs)

**Goal:** Anomaly → RCA summary in DB within 15s. Similarity search returns related incidents.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-045 | LLM infrastructure (`internal/llm/`): `budget.go` (goroutine checks daily spend every 10min, sets Redis `llm:disabled=true` if > `LLM_DAILY_BUDGET_USD`; all LLM callers check this flag first); `cache.go` (3-tier: L1=Redis `rca:{fingerprint}` 30min TTL, L2=Postgres `rca_cache` < 24h, L3=fresh Claude call); `retry.go` (exponential backoff wrapper) | 2h | must | TASK-018 TASK-010 | todo |
| TASK-046 | RCA prompt template (`internal/llm/prompts/rca.tmpl`): Go text/template fields — service, z_score, window_start/end, current_rate, baseline_mean, top_errors ([]struct{Count,Message}), similar_incidents ([]struct{Date,Note}); JSON-mode output schema (summary, likely_cause, confidence, suggested_actions); `rca.tmpl_test.go` renders with sample data and validates JSON schema | 2h | must | — | todo |
| TASK-047 | `internal/rca/` repositories: `repository.go` (EmbeddingRepo impl: Upsert, FindByFingerprint, FindSimilar with pgvector `<=>` cosine distance); `llm_repository.go` (LLMCallRepo impl: Save, SumCostSince) — both GORM with `db.WithContext(ctx)` | 2h | must | TASK-010 | todo |
| TASK-048 | `internal/rca/similarity.go` (SimilarityService): pgvector cosine search top-3, threshold ≥ 0.7, status=RESOLVED, within 90 days; on first sight of new fingerprint: call `LLMClient.Embed` → upsert to `error_embeddings`; budget guard check before every embed call | 2h | must | TASK-047 TASK-045 | todo |
| TASK-049 | `internal/rca/service.go` (RCAService): fingerprint → 3-tier cache check (`internal/llm/cache.go`) → gather context (50 logs ±2min from LogRepo, top-3 error messages, 2 similar from SimilarityService) → render `rca.tmpl` → budget guard check → `LLMClient.Generate` → validate JSON output → persist via `TxManager.RunInTx` (update incident + outbox); rule-based fallback on LLM error/timeout; record LLMCall cost; run `pmctl gen service` | 4h | must | TASK-048 TASK-046 | todo |
| TASK-050 | `internal/rca/consumer.go`: subscribes to `pulse.anomalies`; deserializes `AnomalyDetectedEvent`; calls RCAService; idempotency via dedup; register in consumer supervisor + registry | 1.5h | must | TASK-049 TASK-016 | todo |
| TASK-051 | `internal/rca/handler.go`: `GET /v1/incidents/{id}/rca`; `GET /v1/incidents/{id}/similar`; `GET /v1/insights/cost?range=7d`; `PATCH /v1/incidents/{id}/rca/feedback` (thumbs up/down); JWT + tenant middleware | 1.5h | must | TASK-049 TASK-022 | todo |
| TASK-052 | LLM cron jobs: `internal/cron/cost_recompute_cronjob.go` (daily rollup: SumCostSince(startOfDay), set/reset `llm:disabled`); `internal/cron/retention_cronjob.go` (drop TimescaleDB chunks > 30 days, vacuum audit_logs > 90 days); `internal/cron/outbox_cronjob.go` (delete outbox rows published > 7 days ago); add all to cron command registry | 1h | must | TASK-038 TASK-045 | todo |

**Phase 8 exit criteria:** `bash scripts/break-checkout.sh` → incident fires → RCA summary appears in DB within 15s; `GET /v1/incidents/{id}/similar` returns ≥ 1 result after second demo run; Redis `llm:disabled` absent.

---

## Phase 9: Saga — Incident Resolution (~4 hrs)

**Goal:** Resolve triggers all downstream steps atomically with compensation.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-053 | Incident resolution saga (`internal/saga/`): `saga.go` + `executor.go` (orchestration loop, per-step compensation, CanSkip idempotency); `repository_impl.go` (SagaRepository impl: Load, Save, AdvanceStep, BeginCompensation); `incident_resolution.go` (6-step saga: MarkResolving → GeneratePostmortem → ReclusterEmbeddings → UpdateMetrics → MarkResolved → SendNotification); `steps/` implementations | 2.5h | must | TASK-051 TASK-049 | todo |
| TASK-054 | Update `POST /v1/incidents/{id}/resolve` to trigger saga: validate `{category, note}` (required per BR-13); compute MTTR; launch saga instance via `TxManager.RunInTx`; incident handler returns immediately (saga runs async) | 1.5h | must | TASK-053 TASK-037 | todo |

**Phase 9 exit criteria:** `POST /v1/incidents/{id}/resolve` with `{category, note}` → incident moves through all 6 saga steps; compensations roll back on step failure.

---

## Phase 10: Frontend (~13 hrs)

**Goal:** Open browser, log in, watch live logs, see incidents, read RCA, resolve with note.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-055 | React + Vite + Tailwind scaffold (`web/`): `shadcn/ui`; directory `web/src/{components,pages,hooks,api}`; login page; main layout (sidebar + header); typed API client wrappers | 3h | must | — | todo |
| TASK-056 | Live log table (`web/src/pages/Dashboard.tsx`): WebSocket hook; virtual-scroll (max 500 visible rows); pause/resume toggle; service filter dropdown; exponential backoff reconnect | 4h | must | TASK-044 TASK-055 | todo |
| TASK-057 | Incident list page (`web/src/pages/IncidentList.tsx`): open incident cards (service, severity badge, z-score, detected-at); poll `GET /v1/incidents?status=open` every 10s; WS push updates without reload | 2h | must | TASK-037 TASK-055 | todo |
| TASK-058 | Incident detail page (`web/src/pages/IncidentDetail.tsx`): status badge; timestamps; RCA panel (summary, likely cause, confidence badge, suggested actions, evidence log IDs); similar incidents panel; Acknowledge button; thumbs up/down feedback | 2h | must | TASK-051 TASK-057 | todo |
| TASK-059 | Resolve modal: category dropdown + required note textarea; `POST /v1/incidents/{id}/resolve`; error toast on 4xx/5xx; closes on success | 1h | must | TASK-054 TASK-058 | todo |
| TASK-060 | AI cost insights page (`web/src/pages/Insights.tsx`): Recharts line chart of daily LLM spend from `GET /v1/insights/cost?range=7d` | 1h | should | TASK-051 TASK-055 | todo |

**Phase 10 exit criteria:** `localhost:3000` loads; login with seeded credentials; logs appear within 3s of sim traffic; incident detail shows RCA within 15s of `break-checkout`; resolve modal enforces note.

---

## Phase 11: Polish (~8 hrs)

**Goal:** Demo is smooth end-to-end. Lint clean. Demo script works hands-free.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-061 | `make demo` script (`scripts/demo.sh`): idempotent full reset (truncate + migrate-up + seed) → start sim → trigger `break-checkout` after 30s; prints dashboard URL and walkthrough steps | 1h | must | TASK-039 TASK-013 | todo |
| TASK-062 | Error handling pass: all service errors use `helper/utils/errors/` custom types mapping to gRPC status codes; HTTP handlers return `{code, message}` JSON consistently; React shows toast on 4xx/5xx | 2h | should | — | todo |
| TASK-063 | golangci-lint clean pass: fix all default-config warnings; `//nolint` only where genuinely needed | 1h | must | — | todo |
| TASK-064 | Grafana provisioning (`deploy/grafana/`): "Pulse Internals" (RED metrics per service, Kafka consumer lag, DB connections, Redis memory); "Demo Tenant" (log ingest rate, open incidents, RCA cache hit rate, daily LLM spend) | 1h | should | — | todo |
| TASK-065 | Production `docker-compose.yml` (`deploy/`): Caddy (auto-TLS), Prometheus, Grafana + 4 app services; Docker secrets; `restart: unless-stopped`; `Caddyfile` routes `/v1/logs*` → ingest:8002 and `/v1/*` → api:8000 | 2h | must | — | todo |
| TASK-066 | BadgerDB local ingest buffer (`internal/kafka/buffer.go`): on Kafka publish failure, write to local BadgerDB; background goroutine retries every 5s; clears on success | 1h | should | TASK-015 | todo |

---

## Phase 12: Ship (~13 hrs)

**Goal:** Live on Hetzner VPS with TLS. Benchmarks recorded. README and Loom done.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-067 | Provision Hetzner CCX13 VPS (Ubuntu 24.04): UFW (ports 22/80/443); Docker 24+ + Compose v2; DNS A record for `demo.pulse.dev` | 1h | must | — | todo |
| TASK-068 | Deploy + smoke-test: SSH to VPS; `git clone` + `.env` (mode 0600); `docker compose -f deploy/docker-compose.yml up -d`; `make seed`; verify `https://demo.pulse.dev` loads and `/health` passes | 1h | must | TASK-067 TASK-065 | todo |
| TASK-069 | k6 load tests: `scripts/load-test/ingest-steady.js` (10k logs/sec × 5min); `scripts/load-test/ingest-burst.js` (20k logs/sec × 30s); `scripts/load-test/ws-concurrency.js` (100 concurrent WS); record in `docs/benchmarks/2026-05-xx.md` | 2h | must | TASK-068 | todo |
| TASK-070 | Generate `openapi.yaml` from protos; import to Postman → export `pulse.postman_collection.json`; check into repo | 1h | should | TASK-005 | todo |
| TASK-071 | README: Mermaid architecture diagram (C4 Level 2); animated GIF or screenshot; benchmark table (p99 ingest, p95 RCA, WS capacity); threat model summary; "What I'd build next" list (ClickHouse, Kubernetes, Temporal, Debezium); CI badge | 2h | must | TASK-069 | todo |
| TASK-072 | Helm chart stub (`deploy/helm/`): `Chart.yaml` + `values.yaml` skeleton for api, ingest, worker; `README.helm.md` noting this is future-work scaffold | 0.5h | should | — | todo |
| TASK-073 | Loom demo video (3 min): `make demo` → dashboard turns red → click incident → read RCA → Find Similar → Resolve with note → metrics return normal; add link to README | 1.5h | must | TASK-068 | todo |
| TASK-074 | Blog post draft (1500–2000 words): "LLM Cost Engineering in a Real Go Backend — What Building Pulse Taught Me"; covers tiered model strategy, 3-tier caching, hard budget cap, per-call cost tracking; publish to dev.to or personal site | 2h | should | — | todo |

**Phase 12 exit criteria:** `https://demo.pulse.dev` live; `bash scripts/break-checkout.sh` run remotely → RCA in UI within 15s; k6 confirms p99 < 80ms at 10k logs/sec.

---

## Cut List (trim here first if behind schedule)

1. TASK-042 (Slack smoke test) → skip; trust unit test
2. TASK-060 (Cost insights page) → log to DB, no React UI
3. TASK-064 (Grafana dashboards) → ship Prometheus raw; drop provisioning
4. TASK-066 (BadgerDB buffer) → accept Kafka as hard dependency for demo
5. TASK-072 (Helm chart stub) → drop; mention Kubernetes only in README
6. TASK-074 (Blog post) → write after job interviews
7. TASK-070 (Postman collection) → link to raw OpenAPI YAML instead

---

## Summary

| Phase | Focus | Hours | Tasks |
|---|---|---|---|
| 0 | Foundation (done) | ~18h | TASK-001–007 |
| 1 | Shared infrastructure | ~20h | TASK-008–019 |
| 2 | Reference domain: auth | ~8h | TASK-020–024 |
| 3 | Search domain | ~4h | TASK-025–027 |
| 4 | Ingestion domain | ~10h | TASK-028–033 |
| 5 | Incident domain | ~10h | TASK-034–039 |
| 6 | Notification domain | ~3h | TASK-040–042 |
| 7 | Stream domain | ~5h | TASK-043–044 |
| 8 | RCA domain | ~16h | TASK-045–052 |
| 9 | Saga — resolution | ~4h | TASK-053–054 |
| 10 | Frontend | ~13h | TASK-055–060 |
| 11 | Polish | ~8h | TASK-061–066 |
| 12 | Ship | ~13h | TASK-067–074 |
| **Total** | | **~132h** | **74 tasks** |
