# Pulse — Architecture

Pulse is an AI-native observability platform. It ingests logs over HTTP, detects anomalies via rolling z-score, generates Root Cause Analysis (RCA) using Claude Sonnet, and streams live data via WebSocket.

---

## 1. Five Binaries

Each binary is an independent process with its own config file.

| Binary | Entry point | Port(s) | Purpose |
|---|---|---|---|
| **api** | `cmd/api/main.go` | 8000 (REST) · 8001 (gRPC) | Query logs, manage incidents, trigger RCA, stream events |
| **ingest** | `cmd/ingest/main.go` | 8002 (HTTP) | Accept raw log entries, fingerprint, publish to Kafka |
| **worker** | `cmd/worker/main.go` | — | Kafka consumers: enricher, rca, notify |
| **cron** | `cmd/cron/main.go` | — | Scheduled background jobs (anomaly scan, outbox drain, retention, budget reset) |
| **sim** | `cmd/sim/main.go` | — | Synthetic traffic generator for local demos |

All binaries use `go-zero`'s `service.NewServiceGroup()` to lifecycle-manage goroutines.

---

## 2. Request Flow — Ingest Path

```
Client
  │  POST /v1/ingest/logs
  ▼
ingest binary (8002)
  │  HMAC middleware validates API key signature
  │  Tenant middleware resolves tenant from key hash
  │  Rate-limit middleware (Redis rl:{tenantID}:{window})
  ▼
IngestHandler → IngestService
  │  Compute SHA-256 fingerprint of (service, level, message)
  │  Write LogEntry to Postgres (log_entries hypertable)
  │  Write OutboxEvent in same transaction (ADR-008)
  ▼
Outbox drainer (cron or worker goroutine)
  │  Reads outbox_events WHERE status=pending (FOR UPDATE SKIP LOCKED)
  │  Publishes to Kafka topic: pulse.logs.raw
  │  Deletes outbox row on success
```

---

## 3. Request Flow — Query / API Path

```
Client
  │  REST: GET /v1/logs, /v1/incidents, etc.
  │  gRPC: any RPC method
  ▼
api binary (8000/8001)
  │  JWT auth middleware (Bearer token)
  │  Tenant middleware
  │  Logging + Recovery middleware
  ▼
Handler (internal/{domain}/handler.go)
  │  Validates proto request (protoc-gen-validate)
  │  Calls service method
  ▼
Service (internal/{domain}/service.go)
  │  Business logic, cache checks, LLM calls
  ▼
Repository (internal/{domain}/repository.go) via contract interface
  │  GORM + db.WithContext(ctx)
  ▼
TimescaleDB / Postgres
```

---

## 4. Event-Driven Flow — Kafka Pipeline

```
Kafka topics:
  pulse.logs.raw      ←  ingest binary publishes (via outbox)
  pulse.anomalies     ←  enricher consumer publishes when z-score spike detected
  pulse.rca.request   ←  anomaly or API triggers RCA
  pulse.rca.result    ←  RCA consumer publishes completed analysis
  pulse.outbox        ←  internal outbox event bus

Worker binary subscribes:

  enricher_consumer  (pulse.logs.raw)
    │  EnricherService: compute rolling z-score window
    │  If ZScore > threshold → publish AnomalyDetectedEvent to pulse.anomalies
    │  Store AnomalyEvent in anomaly_events hypertable

  rca_consumer       (pulse.rca.request)
    │  RCAService: check 3-tier cache (L1 Redis → L2 Postgres → L3 Claude)
    │  If cache miss → call LLM budget guard → call Claude Sonnet
    │  Store summary, publish RCACompletedEvent to pulse.rca.result

  notify_consumer    (pulse.rca.result)
    │  NotificationService: send AlertPayload to Slack/PagerDuty
    │  Write AuditLog record
```

### Kafka Event Types (in domain modules)

| Type | File | Topic | Fields |
|---|---|---|---|
| `RawLogEvent` | `internal/ingestion/event.go` | `pulse.logs.raw` | TenantID, ServiceName, Level, Message, Fingerprint, Metadata, Timestamp |
| `AnomalyDetectedEvent` | `internal/incident/event.go` | `pulse.anomalies` | TenantID, ServiceName, Fingerprint, ZScore, WindowSize, LogCount, DetectedAt |
| `RCACompletedEvent` | `internal/rca/event.go` | `pulse.rca.result` | TenantID, IncidentID, Fingerprint, Summary, Model, CompletedAt |

---

## 5. Data Layer

### Postgres / TimescaleDB

| Table | Entity | Notes |
|---|---|---|
| `log_entries` | `LogEntry` | TimescaleDB hypertable, partitioned by `created_at`; composite PK (id, created_at) |
| `anomaly_events` | `AnomalyEvent` | TimescaleDB hypertable, partitioned by `created_at` |
| `incidents` | `Incident` | Standard GORM (soft-delete); status enum, pgtype |
| `error_embeddings` | `ErrorEmbedding` | pgvector column `vector(1536)`; cosine similarity search |
| `llm_calls` | `LLMCall` | Append-only cost ledger; `cost_usd numeric(10,6)` |
| `rca_cache` | `RCACache` | L2 RCA cache; unique on fingerprint |
| `tenants` | `Tenant` | Root multi-tenancy entity |
| `users` | `User` | Tenant-scoped user accounts |
| `api_keys` | `APIKey` | HMAC key hashes; partial unique index (deleted_at IS NULL) |
| `outbox_events` | `OutboxEvent` | Transactional outbox (ADR-008); status: 0=pending, 1=published, 2=failed |
| `saga_instances` | `SagaInstance` | Saga orchestration state; current_step, compensating status |
| `audit_logs` | `AuditLog` | Append-only; no soft-delete; immutable via Postgres RULE |

### Redis

| Key pattern | TTL | Purpose |
|---|---|---|
| `rca:{fingerprint}` | 30 min | L1 RCA cache |
| `llm:disabled` | — | LLM budget guard flag (set by cron when daily cap hit) |
| `rl:{tenantID}:{window}` | 1 min | Per-tenant rate limiting |

---

## 6. LLM Pipeline

### Budget Guard

```
Before any Anthropic API call:
  1. Redis GET llm:disabled
     → if "true": return cached/degraded response, skip LLM
  2. Call LLMClient.Generate / Embed
  3. Record LLMCall (model, tokens, cost_usd) in Postgres
  4. Cron job (midnight UTC): SumCostSince(startOfDay) → if > $5.00, SET llm:disabled
  5. Cron job resets llm:disabled at midnight UTC
```

Daily budget cap: `LLM_DAILY_BUDGET_USD` env var (default `$5.00`).

### Three-Tier RCA Cache

```
RCAService.GetRCA(fingerprint):
  L1: Redis GET rca:{fingerprint}   → HIT: return immediately (30 min TTL)
  L2: Postgres SELECT rca_cache WHERE fingerprint = ? AND created_at > now()-24h
      → HIT: warm L1 and return
  L3: Call Claude Sonnet (after budget guard)
      → Store in L2 (rca_cache upsert) and L1 (Redis SET, 30 min)
      → Return summary
```

---

## 7. Outbox Pattern (ADR-008)

**Rule**: Never publish to Kafka directly inside a DB transaction.

```go
// In any service that does a dual-write:
txManager.RunInTx(ctx, func(tx *gorm.DB) error {
    // 1. Main business write
    incidentRepo.Create(ctx, incident)          // uses tx internally

    // 2. Outbox write IN SAME TRANSACTION
    outboxAppender.Append(ctx, tx, &entity.OutboxEvent{
        Topic:       kafka.TopicAnomalies,
        AggregateID: fingerprint,
        Payload:     json.Marshal(event),
    })
    return nil
})

// Cron / worker drainer:
//   SELECT * FROM outbox_events WHERE status=0 FOR UPDATE SKIP LOCKED LIMIT 100
//   → Publish to Kafka
//   → DELETE row (or mark status=1)
```

Cron job `outbox_cronjob.go` runs the drain loop on a schedule.

---

## 8. Saga Orchestration — Incident Resolution

When an incident is resolved, a saga ensures all downstream steps complete or compensate:

```
Saga: IncidentResolution
  Steps (in order):
    1. MarkResolving      — transition incident status to "resolving"
    2. GeneratePostmortem — call LLM to produce post-mortem summary (with budget guard)
    3. ReclusterEmbeddings — re-embed resolved fingerprint with updated context
    4. UpdateMetrics      — record MTTR and severity metrics
    5. MarkResolved       — transition incident to "resolved", set resolved_at
    6. SendNotification   — deliver resolution alert via Notifier

  Compensation (reverse order on any step failure):
    Each Step implements Compensate() — undoes the forward action
    Each Step implements CanSkip() — idempotent re-run guard

  State persisted in saga_instances:
    current_step (string), status (0=started, 2=compensating, 3=failed)
```

`SagaRepository` (in `internal/saga/repository.go`) tracks step advancement and compensation transitions.

---

## 9. Dependency Injection — Registry

All wiring lives in `internal/registry/`. Handlers and services never construct their own dependencies.

```
┌─────────────────────────────────────────────┐
│              ServiceContext                  │  ← API binary
│  embeds RepositoryContext                    │
│  holds: DB, Redis, ServiceFactory (lazy)     │
│                                              │
│  ServiceFactory (lazy via oncex.OnceValue)   │
│    GetIngestService()   → IngestService      │
│    GetSearchService()   → SearchService      │
│    GetIncidentService() → IncidentService    │
│    GetRCAService()      → RCAService         │
│    GetAuthService()     → AuthService        │
│    ...                                       │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│              ConsumerContext                 │  ← Worker binary
│  embeds RepositoryContext                    │
│  narrower: no HTTP/gRPC layer                │
└─────────────────────────────────────────────┘

┌─────────────────────────────────────────────┐
│               CronContext                    │  ← Cron binary
│  embeds RepositoryContext                    │
│  has access to ServiceFactory for shared svc │
└─────────────────────────────────────────────┘

RepositoryContext:
  Holds all concrete repo instances (GORM-backed)
  Each repo satisfies a contract interface from internal/contract/
```

Services are **lazy-initialized** via `oncex.OnceValue[T]` — created on first access, not at startup. `service_factory.go` is generated by `pmctl gen service` and must not be hand-edited.

---

## 10. Directory Map

`internal/` is organized into **domain modules** — each owns its full vertical slice (handler, service, repository, consumer, mapper, DTO, event). Cross-cutting infrastructure packages stay flat.

```
.
├── api/                        # Proto definitions + generated code (never hand-edit *.pb.go)
│   ├── common/                 #   Shared Pagination, TimestampRange
│   ├── incident/               #   ListIncidents, GetIncident, UpdateStatus, TriggerRCA
│   ├── ingest/                 #   IngestLog
│   ├── search/                 #   SearchLogs, GetLog
│   └── stream/                 #   SubscribeLogs (SSE/WebSocket)
│
├── cmd/                        # Binary entry points (one main.go each)
│   ├── api/
│   ├── cron/
│   ├── ingest/
│   ├── sim/
│   └── worker/
│
├── etc/                        # Config YAML files (${ENV_VAR} substitution)
│
├── external/                   # Third-party HTTP clients (not gRPC generated)
│   ├── anthropic/              #   Claude Sonnet (Generate + Embed)
│   ├── openai/                 #   OpenAI embeddings (fallback)
│   └── slack/                  #   Slack alert delivery
│
├── internal/
│   │
│   │   ── Domain modules (each owns handler + service + repo + consumer + mapper + dto + event)
│   │
│   ├── ingestion/              # Log ingestion pipeline
│   │   ├── handler.go          #   HTTP handler: POST /v1/ingest/logs
│   │   ├── service.go          #   IngestService: fingerprint, write LogEntry + outbox
│   │   ├── enricher.go         #   EnricherService: compute z-score, publish anomaly
│   │   ├── consumer.go         #   Kafka consumer for pulse.logs.raw (enricher)
│   │   ├── repository.go       #   LogRepo implementation (log_entries hypertable)
│   │   ├── server.go           #   Ingest HTTP server builder
│   │   ├── dto.go              #   REST request/response shapes
│   │   ├── mapper.go           #   entity.LogEntry ↔ proto/DTO
│   │   └── event.go            #   RawLogEvent (pulse.logs.raw payload)
│   │
│   ├── incident/               # Incident management + anomaly detection
│   │   ├── handler.go          #   gRPC handler: ListIncidents, GetIncident, UpdateStatus
│   │   ├── service.go          #   IncidentService: CRUD, status transitions
│   │   ├── anomaly.go          #   AnomalyService: rolling z-score detection
│   │   ├── repository.go       #   IncidentRepo implementation (incidents table)
│   │   ├── server.go           #   Incident gRPC server builder
│   │   ├── dto.go              #   REST/DTO shapes
│   │   ├── mapper.go           #   entity.Incident ↔ proto
│   │   └── event.go            #   AnomalyDetectedEvent (pulse.anomalies payload)
│   │
│   ├── rca/                    # Root Cause Analysis
│   │   ├── handler.go          #   gRPC handler: GetRCA, TriggerRCA (insight endpoints)
│   │   ├── service.go          #   RCAService: 3-tier cache → Claude Sonnet
│   │   ├── similarity.go       #   SimilarityService: pgvector cosine search
│   │   ├── consumer.go         #   Kafka consumer for pulse.rca.request
│   │   ├── repository.go       #   EmbeddingRepo implementation (error_embeddings)
│   │   ├── llm_repository.go   #   LLMCallRepo implementation (llm_calls cost ledger)
│   │   ├── mapper.go           #   entity.ErrorEmbedding ↔ proto
│   │   └── event.go            #   RCACompletedEvent (pulse.rca.result payload)
│   │
│   ├── search/                 # Log search
│   │   ├── handler.go          #   gRPC handler: SearchLogs, GetLog
│   │   ├── service.go          #   SearchService: filter + paginate log_entries
│   │   └── dto.go              #   Search request/response shapes
│   │
│   ├── stream/                 # WebSocket live streaming
│   │   ├── handler.go          #   WebSocket upgrade handler
│   │   └── hub.go              #   WebSocket hub (fan-out to connected clients)
│   │
│   ├── auth/                   # Tenant auth, API key management, audit
│   │   ├── handler.go          #   gRPC handler: CreateAPIKey, RevokeAPIKey, ListAPIKeys
│   │   ├── service.go          #   AuthService: tenant + API key business logic
│   │   ├── api_key_repo.go     #   APIKeyRepo implementation (api_keys table)
│   │   ├── tenant_repo.go      #   TenantRepo implementation (tenants table)
│   │   ├── audit_repo.go       #   AuditLogRepo implementation (audit_logs table)
│   │   ├── auth_middleware.go  #   JWT Bearer token validation middleware
│   │   ├── hmac_middleware.go  #   API key HMAC signature validation middleware
│   │   └── tenant_middleware.go #  Resolve tenant from token/key
│   │
│   ├── notification/           # Alerting pipeline
│   │   ├── service.go          #   NotificationService: send AlertPayload, write AuditLog
│   │   └── consumer.go         #   Kafka consumer for pulse.rca.result
│   │
│   │   ── Cross-cutting infrastructure (domain-agnostic)
│   │
│   ├── config/                 # Config structs (one per binary)
│   │
│   ├── consumer/               # Kafka consumer bootstrap/registration hub (worker binary)
│   │   └── consumer.go         #   Registers enricher, rca, notify consumers
│   │
│   ├── contract/               # Interfaces (ports): what services depend on
│   │   ├── repository.go       #   LogRepo, IncidentRepo, EmbeddingRepo, ...
│   │   ├── tx.go               #   TxManager.RunInTx
│   │   ├── outbox.go           #   OutboxAppender, OutboxRepository
│   │   ├── publisher.go        #   KafkaPublisher
│   │   ├── llm.go              #   LLMClient (Generate + Embed)
│   │   ├── cache.go            #   Cache (Get/Set/Del/Exists)
│   │   ├── notifier.go         #   Notifier.SendAlert
│   │   ├── stats.go            #   StatsProvider.ZScore
│   │   └── saga.go             #   Step, Saga, SagaRepository
│   │
│   ├── cron/                   # Scheduled jobs (run in cron binary)
│   │   ├── anomaly_cronjob.go
│   │   ├── outbox_cronjob.go
│   │   ├── retention_cronjob.go
│   │   └── cost_recompute_cronjob.go
│   │
│   ├── handler/                # gRPC + REST route registration only (no domain logic)
│   │   ├── http_handler.go     #   REST gateway route registration (grpc-gateway)
│   │   ├── grpc_handler.go     #   gRPC server registration
│   │   └── health_handler.go   #   Health check endpoint
│   │
│   ├── kafka/                  # Kafka producer/consumer primitives + topic constants
│   │
│   ├── llm/                    # LLM infrastructure: client, budget guard, 3-tier cache, retry, fake
│   │   ├── client.go
│   │   ├── budget.go
│   │   ├── cache.go
│   │   ├── cost_guard.go       #   CostGuardService: daily spend tracking
│   │   ├── retry.go
│   │   └── fake.go
│   │
│   ├── middleware/             # Cross-cutting HTTP/gRPC middleware
│   │   ├── logging_middleware.go
│   │   ├── rate_limit_middleware.go
│   │   └── recovery_middleware.go
│   │
│   ├── outbox/                 # Outbox drainer: publishes outbox_events to Kafka
│   │
│   ├── registry/               # Dependency injection wiring (sole wiring layer)
│   │   ├── base_context.go
│   │   ├── repository_context.go
│   │   ├── service_context.go
│   │   ├── consumer_context.go
│   │   ├── cron_context.go
│   │   └── security_*.go
│   │
│   ├── repository/             # Shared repo implementations (outbox + saga)
│   │   ├── outbox_repository.go
│   │   └── saga_repository.go
│   │
│   ├── saga/                   # Saga orchestration engine
│   │   ├── saga.go
│   │   ├── executor.go
│   │   ├── repository.go
│   │   ├── incident_resolution.go
│   │   └── steps/
│   │       ├── generate_postmortem.go
│   │       ├── mark_resolved.go
│   │       ├── mark_resolving.go
│   │       ├── recluster_embeddings.go
│   │       ├── send_notification.go
│   │       └── update_metrics.go
│   │
│   └── types/
│       ├── define/
│       │   ├── constant/       # Kafka topics, Redis keys, TTLs, budget cap
│       │   └── enum/           # IncidentStatus, OutboxStatus, SagaStatus, Severity
│       ├── entity/             # GORM structs (one per DB table) — shared across domains
│       └── mapper/
│           └── common.go       # Shared mapper helpers (timestamp, pagination, etc.)
│
└── helper/                     # Shared infrastructure wrappers (separate Go module)
    └── utils/
        ├── cache/              #   Redis wrapper
        ├── db/                 #   GORM + primary/replica setup
        ├── authenticator/      #   JWT validation
        ├── authorizer/         #   Permission checks
        └── ...
```

---

## 11. Architectural Rules (ADRs)

| Rule | Where enforced |
|---|---|
| **No LLM calls without budget check** — always verify `llm:disabled` in Redis first | `internal/llm/budget.go` |
| **Outbox for all dual-writes** — never publish to Kafka directly inside a transaction | `contract.OutboxAppender` + `TxManager.RunInTx` |
| **No cross-service DB access** — services call each other via gRPC or Kafka events | Registry context boundaries |
| **Idempotent Kafka consumers** — use upsert, not insert; consumers are safe to replay | `internal/{domain}/consumer.go` |
| **No business logic in handlers** — handlers validate, delegate, return; nothing else | `internal/{domain}/handler.go` |
| **No entity leakage** — handlers see only proto/DTO types; entities stay in service+repo layer | `internal/{domain}/mapper.go` |
| **Interface before implementation** — define in `internal/contract/`, inject via constructor | `internal/registry/` |
| **WithContext on every GORM call** — `db.WithContext(ctx)` is mandatory | `internal/{domain}/repository.go` |
| **Mappers are the only conversion point** — entity ↔ proto mapping belongs only in the domain's `mapper.go` | Enforced by code review |

---

## 12. Infrastructure Stack

| Component | Technology | Purpose |
|---|---|---|
| Primary DB | Postgres 15 + TimescaleDB + pgvector | Log storage, incident data, vector search |
| Cache | Redis 7 | RCA L1 cache, LLM budget flag, rate limiting |
| Message broker | Kafka | Async event pipeline between binaries |
| LLM | Anthropic Claude Sonnet (primary) · OpenAI (fallback) | RCA generation, error embeddings |
| Alerts | Slack | Operational alert delivery |
| Auth | JWT (Bearer) + HMAC API keys | Dual auth model for users vs. services |
| HTTP/gRPC | go-zero + grpc-gateway | Unified REST+gRPC from single proto definition |
