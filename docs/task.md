# Pulse — Sprint Task List

**Project:** Pulse — AI-Native Observability Platform (Portfolio Edition)
**Total budget:** ~150 hours over 12 weeks (~2.5 hrs/day, 5 days/week)
**Last updated:** 2026-05-11

---

## How to Read This File

| Field | Meaning |
|---|---|
| **ID** | Unique task identifier (`TASK-NNN`) |
| **Est** | Estimated hours |
| **Pri** | `must` = non-negotiable for demo, `should` = cut if over budget |
| **Deps** | Tasks that must complete first |
| **Status** | `todo` / `in-progress` / `done` |

The **non-negotiable core** is: ingestion → anomaly → RCA → dashboard. Everything else is `should`.

---

## Sprint 1–2: Foundation (~27 hrs, Weeks 1–2)

**Goal:** `make up` brings three healthy services. Proto contracts generated. CI pipeline is green.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-001 | Init Go monorepo from `tronglv-pulse` boilerplate: set `go.mod` module name to `pulse`; create `cmd/api`, `cmd/ingest`, `cmd/worker`, `cmd/cron`, `cmd/sim`; set up `go.work` (workspace includes `helper/` module); add `pmctl.gen.yaml`, `CLAUDE.md`, `.air.toml`, `.mcp.json`, `.env.example`, `.gitignore` | 2h | must | — | todo |
| TASK-002 | Dev `docker-compose.yml`: Postgres 16 (`timescale/timescaledb-ha:pg16-all`), Redis 7, Redpanda (KRaft single-node), pgAdmin 4; all on `pulse` bridge network; bind port Postgres `:5433`, Redis `:6379`, Redpanda `:9092`, pgAdmin `:5050` | 2h | must | TASK-001 | todo |
| TASK-003 | DB migrations 0001–0009 (`db/migrations/`): write both `.up.sql` and `.down.sql` for every migration — extensions (timescaledb, vector, pg_trgm, citext), logs hypertable, `logs_metrics_5min` continuous aggregate, incidents, error_embeddings, llm_calls, audit_logs, outbox_events, saga_instances | 5h | must | TASK-002 | todo |
| TASK-004 | GORM entities (`internal/types/entity/`): `Log`, `Incident`, `ErrorEmbedding`, `LLMCall`, `AuditLog`, `Tenant`, `User`, `APIKey`, `OutboxEvent`, `SagaInstance`; register all in `migrate.go`; add `internal/types/define/constant/constant.go` (topic names, cache TTLs, rate limits) + `permission.go`; add `internal/types/define/enum/enum.go` (Severity, IncidentStatus, ResolutionCategory); add `internal/types/event/` Kafka payload structs (`RawLogEvent`, `AnomalyEvent`, `RCACompletedEvent`) | 3.5h | must | TASK-003 | todo |
| TASK-005 | Proto contracts: write `.proto` files for all 5 domains in `api/` (`ingest.proto`, `incident.proto`, `search.proto`, `stream.proto`, `common.proto`); vendor shared proto deps into `protos/` (google/api, google/protobuf, google/rpc, validate, openapi/v3, protoc-gen-openapiv2); run `make grpc && make validate && make gateway` → verify all generated files appear (`*.pb.go`, `*_grpc.pb.go`, `*_http.pb.go`, `*.pb.validate.go`) | 3h | must | TASK-001 | todo |
| TASK-006 | Port interfaces (`internal/contract/`): write all 9 interface files — `repository.go` (LogRepo, IncidentRepo, EmbeddingRepo, LLMCallRepo, TenantRepo, APIKeyRepo, OutboxRepo, SagaRepo, AuditLogRepo), `publisher.go` (KafkaPublisher), `outbox.go` (OutboxAppender + OutboxRepository), `saga.go` (Step + Saga + SagaRepository), `tx.go` (TxManager / RunInTx), `llm.go` (LLMClient: Generate + Embed), `notifier.go` (Notifier: SendAlert), `cache.go` (Cache: Get/Set/Del), `stats.go` (StatsProvider: ZScore) | 2h | must | TASK-004 | todo |
| TASK-007 | Config structs (`internal/config/config.go`): single Config struct per binary; write `etc/api.yaml`, `etc/ingest.yaml`, `etc/worker.yaml`, `etc/cron.yaml` using `${ENV_VAR}` substitution; complete `.env.example` with all variables and defaults | 1.5h | must | TASK-005 | todo |
| TASK-008 | `GET /health` endpoint on each of the three services (no auth, no DB call); verify with `curl localhost:{8000,8002}/health` | 1h | must | TASK-007 | todo |
| TASK-009 | `internal/registry/` DI composition root (mirrors boilerplate exactly): `base_context.go` (BaseContext interface: GetConfig, GetDownloader), `repository_context.go` (RepositoryContext interface + impl: all repo getters), `service_context.go` (ServiceContext interface + impl: HTTP/gRPC wiring), `consumer_context.go` (ConsumerContext + impl: Kafka consumer wiring), `cron_context.go` (CronContext + impl: scheduled job wiring), `security_authentication.go`, `security_authorization.go`, `security_http.go`, `security_context.go` | 3h | must | TASK-006 | todo |
| TASK-010 | Makefile targets: `make up`, `make down`, `make run-api`, `make run-ingest`, `make run-worker`, `make run-cron`, `make air`, `make test`, `make coverage`, `make grpc`, `make validate`, `make gateway`, `make mocks`, `make migrate-up`, `make migrate-down`, `make seed`, `make demo` | 1.5h | must | TASK-001 | todo |
| TASK-011 | GitHub Actions CI workflow (`.github/workflows/ci.yml`): `golangci-lint run`, `go test -race ./...`, `go test -bench=. -benchmem ./cmd/ingest/...`, `docker build` for each of the 4 Dockerfiles; push images to `ghcr.io` on tag; add CI badge to README | 2h | must | TASK-001 | todo |
| TASK-012 | Seed script (`scripts/seed.sh`): insert demo tenant, demo user (bcrypt cost-12 hashed pw), demo API key (`pk_demo_xxx`) with known HMAC secret; idempotent (upsert, not insert) | 1h | should | TASK-003 | todo |

**Sprint exit criteria:** `make up` starts three services; all `/health` endpoints return 200; `make grpc` regenerates without errors; `make test` passes; CI pipeline is green on push.

---

## Sprint 3–4: Ingestion Pipeline (~25 hrs, Weeks 3–4)

**Goal:** Logs flow end-to-end from HTTP POST → Kafka → TimescaleDB. Visible by querying the DB.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-013 | `internal/handler/http_handler.go`: `RegisterHTTPHandlers` mounts all REST routes and attaches middleware chain; `internal/handler/ingest_handler.go`: implements `POST /v1/logs` (delegates to ingest service); `internal/handler/health_handler.go`: `/health` + `/v1/metrics` (Prometheus) | 2h | must | TASK-009 | todo |
| TASK-014 | HMAC middleware (`internal/middleware/hmac_middleware.go`): verify `X-Signature: sha256=<hmac-of-body>` against `INGEST_HMAC_SECRET`; 401 on failure; reuse `helper/utils/toolkit/crypto.go` from boilerplate; add `helper/hmac/hmac.go` thin wrapper | 1.5h | must | TASK-013 | todo |
| TASK-015 | API key auth middleware (`internal/middleware/auth_middleware.go`): validate `Authorization: Bearer pk_xxx`; lookup prefix → hash in `api_keys` table, cache hit/miss via `helper/utils/cache/`; 401 on failure | 1.5h | must | TASK-013 | todo |
| TASK-016 | Rate limiter middleware (`internal/middleware/rate_limit_middleware.go`): Redis token-bucket, 50k logs/sec per tenant; 429 + `Retry-After` on exceed; reuse `helper/utils/stores/redis/` | 1.5h | must | TASK-013 | todo |
| TASK-017 | Request DTO + fingerprint: `internal/types/dto/ingest_dto.go` (LogEntry, LogBatch); `helper/fingerprint/fingerprint.go` — MD5(service + top-error-message + severity) → CHAR(32) | 1h | must | TASK-004 | todo |
| TASK-018 | Kafka producer (`internal/kafka/producer.go`): wraps `segmentio/kafka-go`; idempotent batching; `acks=all`; publishes to `logs.raw`; `internal/kafka/topics.go` (all 4 topic name constants) | 2h | must | TASK-009 | todo |
| TASK-019 | Ingest service (`internal/service/ingest_service.go`): validate batch (≤500 entries, ≤2MB), compute fingerprint per entry, call `KafkaPublisher.Publish(logs.raw)`; return accepted_count; run `pmctl gen service` after → regenerates `service_factory.go` | 2h | must | TASK-017 TASK-018 | todo |
| TASK-020 | Ingest repository (`internal/repository/api_key_repository.go`, `internal/repository/log_repository.go`): API key lookup (prefix → hash compare); insert to TimescaleDB hypertable via GORM; both implement their `internal/contract/repository.go` interfaces | 2h | must | TASK-009 | todo |
| TASK-021 | BadgerDB local buffer (`internal/kafka/buffer.go`): on `Kafka.Publish` failure, write to local BadgerDB; background goroutine retries every 5s; clears on success | 2.5h | should | TASK-018 | todo |
| TASK-022 | Kafka consumer scaffold (`internal/kafka/consumer.go`): wraps `kq.MustNewQueue`; `ConsumeHandler` interface; `internal/consumer/consumer.go` supervisor — one goroutine per consumer with panic-recovery restart | 2h | must | TASK-009 | todo |
| TASK-023 | Enricher consumer (`internal/consumer/enricher_consumer.go`): subscribes to `logs.raw`; calls enricher service; publishes to `logs.enriched` via Outbox; `internal/service/enricher_service.go` — parse, fingerprint, GORM insert to logs hypertable | 3h | must | TASK-022 | todo |
| TASK-024 | Synthetic log generator (`cmd/sim/main.go`): HTTP POST to `/v1/logs` with valid HMAC; flags `--target`, `--api-key`, `--hmac-secret`, `--service`, `--rps`, `--error-rate`, `--duration`; `scripts/simulate-traffic.sh` runs at 100 rps for 60s | 2h | must | TASK-013 | todo |

**Sprint exit criteria:** `bash scripts/simulate-traffic.sh` runs for 60 seconds; `SELECT count(*) FROM logs` in Postgres returns > 5000 rows.

---

## Sprint 5–6: Realtime Dashboard (~25 hrs, Weeks 5–6)

**Goal:** Open browser, login, watch logs streaming live in the log table.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-025 | Auth service + handler: `internal/service/auth_service.go` (bcrypt compare via `helper/utils/authenticator/`, issue 15m access token + 30d refresh token, rotation on refresh); `internal/handler/auth_handler.go` (`POST /v1/auth/login`, `POST /v1/auth/refresh`); `internal/repository/tenant_repository.go` (user lookup by email) | 3h | must | TASK-009 | todo |
| TASK-026 | Auth + tenant middleware: `internal/middleware/auth_middleware.go` (JWT verify via `helper/utils/authenticator/`); `internal/middleware/tenant_middleware.go` (extract tenant_id from JWT claims → context, reuse `helper/utils/identity/`); `internal/middleware/recovery_middleware.go` (panic → 500, reuse `helper/utils/recovery/`) | 1.5h | must | TASK-025 | todo |
| TASK-027 | WebSocket hub (`internal/server/ws_hub.go`): in-process pub/sub; browser WS connection registers a typed channel with service filter; Kafka events fan out to matching connections; sample beyond 1000 msgs/sec per client; thread-safe with `sync.RWMutex` | 3h | must | TASK-009 | todo |
| TASK-028 | WebSocket handler (`internal/handler/stream_handler.go`): `GET /v1/stream?token=<jwt>`; JWT validation; read subscribe message `{action, channel, filter}`; write `{type, data}` frames; heartbeat ping every 30s; auto-close on auth expiry | 2h | must | TASK-027 | todo |
| TASK-029 | Kafka → WS bridge (`internal/consumer/`): consumer subscribes to `logs.enriched` + `anomalies.detected` + `rca.completed`; pushes deserialized events to `ws_hub` channels; add to `consumer.go` supervisor | 2h | must | TASK-027 TASK-023 | todo |
| TASK-030 | Incident + search handlers: `internal/handler/incident_handler.go` (`GET /v1/incidents`, `GET /v1/incidents/{id}`); `internal/handler/search_handler.go` (`GET /v1/logs/search?q=&service=&from=&to=`); `internal/handler/grpc_handler.go` `RegisterGRPCHandlers`; `internal/repository/incident_repository.go` (GORM CRUD) | 2.5h | must | TASK-009 | todo |
| TASK-031 | Incident + search services: `internal/service/incident_service.go` (list + get; ack/resolve stubs for later); `internal/service/search_service.go` (full-text via `pg_trgm`); `internal/types/dto/incident_dto.go`, `internal/types/dto/search_dto.go`; run `pmctl gen service` | 2h | must | TASK-030 | todo |
| TASK-032 | React + Vite + Tailwind scaffold (`web/`): init with `shadcn/ui`; directory structure `web/src/{components,pages,hooks,api}`; login page; main layout (sidebar + header); typed API client wrappers in `web/src/api/` | 3h | must | — | todo |
| TASK-033 | Live log table (`web/src/pages/Dashboard.tsx`): WebSocket hook (`web/src/hooks/useWebSocket.ts`); virtual-scroll table (max 500 visible rows); pause/resume toggle; service filter dropdown; reconnect with exponential backoff on WS disconnect | 4h | must | TASK-028 TASK-032 | todo |
| TASK-034 | Incident list page (`web/src/pages/IncidentList.tsx`): open incident cards with service name, severity badge, z-score, detected-at; poll `GET /v1/incidents?status=open` every 10s; WS push updates list without page reload; empty + loading states | 2h | must | TASK-031 TASK-032 | todo |

**Sprint exit criteria:** `localhost:3000` loads; login succeeds with seeded credentials; live logs appear in table within 3 seconds of sim traffic; incident list page renders.

---

## Sprint 7–8: Anomaly Detection (~25 hrs, Weeks 7–8)

**Goal:** `make break-checkout` → incident appears in UI within 60 seconds.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-035 | Z-score + EWMA utilities: `helper/stats/zscore.go` (rolling z-score; configurable baseline window; sparse-data fallback returns 0); `helper/stats/ewma.go` | 2h | must | — | todo |
| TASK-036 | `logs_metrics_5min` continuous aggregate (migration TASK-003): verify query returns per-service error rates via psql; confirm TimescaleDB refresh policy fires; add a manual `CALL refresh_continuous_aggregate(...)` to seed script | 1h | must | TASK-003 TASK-012 | todo |
| TASK-037 | Cron command registry (`internal/cron/command.go`): mirrors boilerplate pattern — CLI command map keyed by job name; `cmd/cron/main.go` reads `CRON_JOB` env var and dispatches; makes each cron job independently runnable | 1h | must | TASK-009 | todo |
| TASK-038 | Anomaly detector cron job (`internal/cron/anomaly_cronjob.go`): ticker every 10s; queries `logs_metrics_5min` via GORM; computes z-score per (tenant, service) using `helper/stats/zscore.go`; 10-minute cooldown per `(service, metric)` in Redis; sparse-data fallback (error_rate > 5%); calls `anomaly_service.Fire()` on trigger | 4h | must | TASK-035 TASK-036 TASK-037 | todo |
| TASK-039 | Anomaly service + incident persistence (`internal/service/anomaly_service.go`): validate z-score threshold; insert `Incident` (status=OPEN, z_score, fingerprint, severity) within a `TxManager.RunInTx`; append `OutboxEvent` in same transaction; run `pmctl gen service` | 2.5h | must | TASK-038 TASK-009 | todo |
| TASK-040 | Outbox publisher (`internal/outbox/publisher.go`): goroutine in `cmd/worker`; `FOR UPDATE SKIP LOCKED` query fetches up to 100 unpublished rows every 500ms; publishes each to Kafka via `KafkaPublisher`; exponential backoff (max 5 min) on failure; `MarkPublished` on success; `internal/repository/outbox_repository.go` implements `contract.OutboxRepository` | 3h | must | TASK-039 | todo |
| TASK-041 | Outbox cleanup cron job (`internal/cron/outbox_cronjob.go`): daily job deletes `outbox_events WHERE published_at IS NOT NULL AND created_at < NOW() - INTERVAL '7 days'`; add to cron command registry | 0.5h | must | TASK-037 TASK-040 | todo |
| TASK-042 | Consumer-side idempotency (`internal/outbox/dedup.go`): Redis SET keyed by `event_id` with 1h TTL; all consumers call `dedup.IsSeen(event_id)` before processing; reuse `helper/utils/stores/redis/` | 1h | must | TASK-040 | todo |
| TASK-043 | Incident lifecycle endpoints: `internal/handler/incident_handler.go` adds `POST /v1/incidents/{id}/ack` + `POST /v1/incidents/{id}/resolve`; `internal/service/incident_service.go` handles status transitions + MTTR calculation via `contract.TxManager`; `internal/repository/incident_repository.go` adds `UpdateStatusTx` and `UpdateRCATx` (for Outbox composition) | 2.5h | must | TASK-031 TASK-040 | todo |
| TASK-044 | Incident detail page (`web/src/pages/IncidentDetail.tsx`): status badge (OPEN=red, ACK=yellow, RESOLVED=green); z-score display; detected/acknowledged/resolved timestamps; Acknowledge button; WS push updates status badge without reload | 2h | must | TASK-043 TASK-034 | todo |
| TASK-045 | WS push for incidents: `ws_hub` receives `anomalies.detected` events from bridge (TASK-029); fan-outs to dashboard; React dashboard updates incident count badge and shows new incident card without manual refresh | 1h | must | TASK-029 TASK-044 | todo |
| TASK-046 | Demo scripts: `scripts/break-checkout.sh` (runs sim at `--error-rate=0.5 --service=checkout --duration=30s`); `scripts/fix-checkout.sh` (reverts checkout to `--error-rate=0.01`) | 1h | must | TASK-024 | todo |

**Sprint exit criteria:** `bash scripts/break-checkout.sh` → wait 60s → incident appears in UI with status OPEN and z-score displayed; click Acknowledge → status badge changes to ACK.

---

## Sprint 9–10: AI Integration (~24 hrs, Weeks 9–10)

**Goal:** Anomaly → RCA visible in UI within 15 seconds. Find-similar works.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-047 | Anthropic client (`external/anthropic/client.go`): wraps `anthropic-ai/sdk-go`; structured output via `tool_choice`; 20s timeout; retry once on JSON parse failure; implements `contract.LLMClient.Generate()`; `external/anthropic/dto.go` (request/response shapes) | 3h | must | TASK-006 | todo |
| TASK-048 | OpenAI embedding client (`external/openai/client.go`): calls `text-embedding-3-small`; implements `contract.LLMClient.Embed()`; returns `[]float32` (1536 dims); `external/openai/dto.go` | 1.5h | must | TASK-006 | todo |
| TASK-049 | LLM fake client (`internal/llm/fake.go`): deterministic output keyed by input hash; used in all unit tests; no network calls; implements full `contract.LLMClient` interface | 1h | must | TASK-047 | todo |
| TASK-050 | Enricher service update (`internal/service/enricher_service.go`): on first sight of a new fingerprint (not in `error_embeddings`), call `LLMClient.Embed(top_error_message)` → upsert to `error_embeddings` with pgvector column; `internal/repository/embedding_repository.go` implements `EmbeddingRepo.UpsertEmbedding()` and `CosineSimilaritySearch()` | 2.5h | must | TASK-047 TASK-048 | todo |
| TASK-051 | RCA prompt template (`internal/llm/prompts/rca.tmpl`): Go template fields — service, anomaly_type, z_score, window_start/end, current_rate, baseline_mean, top_errors ([]struct{Count, Message}), similar_incidents ([]struct{Date, ResolutionNote}); JSON-mode output schema (summary, likely_cause, confidence, suggested_actions); `rca.tmpl_test.go` renders with sample data and verifies JSON structure | 2h | must | TASK-047 | todo |
| TASK-052 | 3-tier RCA cache (`internal/llm/cache.go`): L1 = Redis `rca:{fingerprint}` TTL 30m (hit → return, mark `cached:true`); L2 = Postgres `incidents.rca_summary` WHERE same fingerprint AND resolved_at > NOW()-24h (hit → refresh L1, return); L3 = fresh Sonnet call → write both L1 and L2 | 2.5h | must | TASK-051 | todo |
| TASK-053 | RCA service (`internal/service/rca_service.go`): fingerprint → cache.Check() → gather context (50 logs ±2min from log_repository, top-3 error messages, 2 similar from EmbeddingRepo); render rca.tmpl; call `LLMClient.Generate()`; validate JSON; persist via `TxManager.RunInTx(UpdateRCATx + OutboxAppend)`; rule-based fallback on LLM error/timeout; run `pmctl gen service` | 4h | must | TASK-052 TASK-043 | todo |
| TASK-054 | LLM cost tracking + budget guard: `internal/repository/llm_call_repository.go` inserts every Sonnet call (input_tokens, output_tokens, cost_usd, latency_ms, cache_hit); `internal/llm/budget.go` goroutine recalculates daily sum every 10 min; sets Redis `llm:disabled=true` if > $5; RCA path checks flag first; `internal/cron/cost_recompute_cronjob.go` daily rollup + add to command registry | 2h | must | TASK-053 | todo |
| TASK-055 | Similar incidents: `internal/service/similarity_service.go` (pgvector cosine search top-3, similarity > 0.7, status=RESOLVED, within 90 days via `EmbeddingRepo.CosineSimilaritySearch()`); `GET /v1/incidents/{id}/similar` → `internal/handler/incident_handler.go`; `internal/types/dto/incident_dto.go` adds `SimilarIncidentResponse` | 2h | must | TASK-050 | todo |
| TASK-056 | RCA consumer (`internal/consumer/rca_consumer.go`): subscribes to `anomalies.detected`; calls `rca_service.Generate()`; checks idempotency via `dedup.IsSeen(event_id)`; add to consumer supervisor | 1.5h | must | TASK-053 TASK-042 | todo |
| TASK-057 | Incident detail page — RCA panel (`web/src/pages/IncidentDetail.tsx` update): RCA summary text; likely cause; confidence badge (low=grey, medium=yellow, high=green); suggested actions ordered list; evidence log IDs as clickable links; thumbs up/down feedback stored via `PATCH /v1/incidents/{id}/rca/feedback` | 2h | must | TASK-053 TASK-044 | todo |
| TASK-058 | Similar incidents panel (`web/src/pages/IncidentDetail.tsx` update): renders alongside RCA panel; shows past incident date, service, MTTR, resolution note | 1h | should | TASK-055 TASK-057 | todo |
| TASK-059 | AI cost insights: `GET /v1/insights/cost?range=7d` → `internal/handler/insight_handler.go`; React page `web/src/pages/Insights.tsx` with single Recharts line chart of daily spend | 1h | should | TASK-054 | todo |

**Sprint exit criteria:** `bash scripts/break-checkout.sh` → incident fires → RCA appears in UI within 15 seconds with non-empty summary and ≥1 suggested action; "Find similar" returns ≥1 result after second demo run; Redis `llm:disabled` flag is absent.

---

## Sprint 11: Polish (~12 hrs, Week 11)

**Goal:** Demo is smooth end-to-end. Slack alert fires. Resolve flow enforces required note.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-060 | Slack notifier: `external/slack/client.go` (POST to webhook URL; uses `helper/utils/httpc/` for retry × 5 with backoff; implements `contract.Notifier`); `internal/service/notification_service.go` (formats message: service, z-score, RCA summary, link); notify consumer (`internal/consumer/notify_consumer.go`) subscribes to `rca.completed` | 2h | should | TASK-053 TASK-042 | todo |
| TASK-061 | Incident resolution flow update: `POST /v1/incidents/{id}/resolve` validates `{category, note}` (both required — BR-13); computes MTTR (resolved_at - detected_at); updates status=RESOLVED via `TxManager.RunInTx(UpdateStatusTx + OutboxAppend)`; saga trigger stub (full saga in TASK-062) | 1.5h | must | TASK-043 | todo |
| TASK-062 | Incident resolution saga (`internal/saga/`): `saga.go` (Step interface with Execute/Compensate/CanSkip), `executor.go` (orchestration loop + per-step compensation), `repository.go` (saga_instances CRUD via `internal/repository/saga_repository.go`), `incident_resolution.go` (6-step saga definition), `steps/` (mark_resolving, generate_postmortem, recluster_embeddings, update_metrics, send_notification, mark_resolved); `mark_resolved` uses Outbox internally | 2.5h | must | TASK-061 TASK-053 | todo |
| TASK-063 | Resolve button in React (`web/src/pages/IncidentDetail.tsx`): modal with category dropdown (deploy issue / code bug / external / infra / other) + required note textarea; submit calls `POST /v1/incidents/{id}/resolve`; closes modal on success; shows error toast on failure | 1.5h | must | TASK-061 TASK-057 | todo |
| TASK-064 | `make demo` script (`scripts/demo.sh`): idempotent full reset (truncate logs + incidents + embeddings + outbox + saga_instances) → `make migrate-up` → `make seed` → start sim → trigger `break-checkout` after 30s; prints dashboard URL and demo walkthrough steps | 1h | must | TASK-046 TASK-012 | todo |
| TASK-065 | Error handling pass: all service errors wrap using `helper/utils/errors/` custom types (map to gRPC status codes); HTTP handlers return `{code, message}` JSON consistently via `helper/response/response.go`; React shows toast on 4xx/5xx; `internal/middleware/logging_middleware.go` logs structured request + response | 2h | should | — | todo |
| TASK-066 | Data retention cron job (`internal/cron/retention_cronjob.go`): daily job drops TimescaleDB chunks older than 30 days; vacuum `audit_logs` older than 90 days; add to command registry | 0.5h | should | TASK-037 | todo |
| TASK-067 | golangci-lint clean pass: fix all default-config warnings across all packages; add `//nolint:xxx // reason` only where genuinely needed | 1h | must | — | todo |
| TASK-068 | Grafana dashboards: provision JSON dashboards in `deploy/grafana/provisioning/dashboards/` and `deploy/grafana/provisioning/datasources/`; "Pulse Internals" (RED metrics per service, Kafka consumer lag, DB active connections, Redis memory); "Demo Tenant" (log ingestion rate, open incident count, RCA cache hit rate, daily LLM spend) | 1h | should | — | todo |

**Sprint exit criteria:** Full 5-minute demo (`make demo`) executes without errors; Slack message arrives within 5s of RCA completion; Resolve modal enforces note; lint is clean.

---

## Sprint 12: Ship (~13 hrs, Week 12)

**Goal:** Live on Hetzner VPS with TLS. Benchmark recorded. README, Postman collection, and Loom done.

| ID | Task | Est | Pri | Deps | Status |
|---|---|---|---|---|---|
| TASK-069 | Provision Hetzner CCX13 VPS (Ubuntu 24.04): configure UFW (ports 22/80/443 only); install Docker 24+ + Docker Compose v2; configure DNS A record for `demo.pulse.dev` | 1h | must | — | todo |
| TASK-070 | Production `deploy/docker-compose.yml`: Caddy (auto-TLS), Prometheus, Grafana + 4 app services (api, ingest, worker, cron); Docker secrets for Postgres password; all services `restart: unless-stopped`; `deploy/Caddyfile` reverse-proxies `/v1/logs*` → ingest:8002 and `/v1/*` → api:8000 | 2h | must | TASK-069 | todo |
| TASK-071 | Deploy and smoke-test: SSH to VPS; `git clone` + set `.env` (mode 0600); `docker compose -f deploy/docker-compose.yml up -d`; run `make seed`; open `https://demo.pulse.dev`; confirm live logs and health check passes | 1h | must | TASK-070 | todo |
| TASK-072 | k6 load tests: run `scripts/load-test/ingest-steady.js` (10k logs/sec × 5 min); `scripts/load-test/ingest-burst.js` (20k logs/sec × 30s); `scripts/load-test/ws-concurrency.js` (100 concurrent WS); record results in `docs/benchmarks/2026-05-xx.md` with p99 latency + Grafana screenshots | 2h | must | TASK-071 | todo |
| TASK-073 | Generate `openapi.yaml` from protos (protoc-gen-openapiv2); import into Postman → export as `pulse.postman_collection.json`; check into repo; verify all 12 endpoints are present with correct request schemas | 1h | should | TASK-005 | todo |
| TASK-074 | README: Mermaid architecture diagram (matches C4 Level 2 from Section 4); animated demo GIF or screenshot; benchmark numbers table (p99 ingest, p95 RCA, WS capacity); threat model summary; links to all 9 ADRs in `docs/adr/`; "What I'd build next" list (ClickHouse, Kubernetes, Temporal, Debezium); CI badge | 2h | must | TASK-072 | todo |
| TASK-075 | Stub Helm chart (`deploy/helm/`): `Chart.yaml` + `values.yaml` skeleton for api, ingest, worker services; one-liner `README.helm.md` explaining this is a future-work scaffold; does not need to deploy | 0.5h | should | — | todo |
| TASK-076 | Loom demo video (3 minutes): `make demo` → dashboard turns red → click incident → read RCA → click Find Similar → click Resolve (with note) → watch metrics return to normal; upload and add link to README | 1.5h | must | TASK-071 | todo |
| TASK-077 | Blog post draft (1500–2000 words): "LLM Cost Engineering in a Real Go Backend — What Building Pulse Taught Me"; covers tiered model strategy, 3-tier caching, hard budget cap, per-call cost tracking; publish to dev.to or personal site | 2h | should | — | todo |

**Sprint exit criteria:** Interviewer opens `https://demo.pulse.dev`, sees live dashboard; `bash scripts/break-checkout.sh` run remotely, RCA appears within 15 seconds; k6 confirms p99 < 80ms at 10k logs/sec.

---

## Cut List (in order — trim here first if behind schedule)

1. TASK-060 (Slack) → show in-UI notification only; cut external/slack entirely
2. TASK-058 (Find-similar panel) → keep endpoint, drop React UI; show JSON in browser
3. TASK-059 (Cost insights page) → log to DB, no React UI
4. TASK-077 (Blog post) → write after job interviews
5. TASK-075 (Helm chart stub) → drop; mention Kubernetes only in README
6. TASK-068 (Grafana dashboards) → ship Prometheus raw; drop Grafana provisioning
7. TASK-021 (BadgerDB buffer) → accept Kafka as hard dependency for demo

The **non-negotiable core chain**:
`TASK-001 → TASK-003 → TASK-005 → TASK-019 → TASK-023 → TASK-033 → TASK-039 → TASK-053 → TASK-061 → TASK-071 → TASK-074`

---

## Summary

| Sprint | Focus | Hours | Task range |
|---|---|---|---|
| 1–2 | Foundation + Proto contracts | ~27h | TASK-001–012 |
| 3–4 | Ingestion pipeline | ~25h | TASK-013–024 |
| 5–6 | Realtime dashboard | ~25h | TASK-025–034 |
| 7–8 | Anomaly detection | ~25h | TASK-035–046 |
| 9–10 | AI integration | ~24h | TASK-047–059 |
| 11 | Polish | ~12h | TASK-060–068 |
| 12 | Ship | ~13h | TASK-069–077 |
| **Total** | | **~151h** | **77 tasks** |
