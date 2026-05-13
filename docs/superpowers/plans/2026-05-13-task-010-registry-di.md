# TASK-010: Registry DI Composition Root Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** Implement `internal/registry/` as the sole DI wiring layer for all four binary entry points (api, ingest, worker, cron) — matching the pattern established in `tronglv-pulse/internal/registry/`.

**Architecture:** Each binary has its own typed context (`ServiceContext`, `IngestContext`, `ConsumerContext`, `CronContext`) that composes a common `BaseContext` and optionally `RepositoryContext`, `SecurityContext`, and `ServiceFactory`. All repository getters return `nil` stubs now — concrete implementations land in TASK-011+. All security middleware is passthrough — real JWT/HMAC wiring lands in the auth tasks.

**Tech Stack:** go-zero v1.9.2 · GORM + `gorm.io/driver/postgres` · `oncex.OnceValue[T]` for lazy init · `github.com/redis/go-redis/v9` (future) · `contract.*` interfaces · `downloader.NewDownloader()`

---

## Reference Pattern (tronglv-pulse)

The `tronglv-pulse` project at `tronglv-pulse/internal/registry/` is the gold standard. Key rules:
- Every context embeds `BaseContext` (GetDownloader)
- `ServiceContext` and `CronContext` both vend `*service.ServiceFactory` via `oncex.OnceValue`
- `RepositoryContext` is a concrete struct wrapping `*gorm.DB` — repo getters are all stubs initially
- `SecurityContext` composes three sub-contexts: Authentication, Authorization, HttpSecurity
- `NewServiceContext` opens the DB and runs AutoMigrate at startup ("fail fast")

Key difference from tronglv-pulse: this repo has **four different config types** (APIConfig, IngestConfig, WorkerConfig, CronConfig). `ServiceFactoryContext` therefore does NOT include `GetConfig()` — services receive dependencies directly.

---

## File Map

| File | Action | Responsibility |
|---|---|---|
| `internal/registry/base_context.go` | create | `BaseContext` interface + `baseContext` struct |
| `internal/registry/db.go` | create | `mustOpenDB(dsn)` — shared GORM opener |
| `internal/registry/repository_context.go` | create | `RepositoryContext` interface + nil-stub struct |
| `internal/registry/security_authentication.go` | create | `AuthenticationContext` — JWT secret accessor |
| `internal/registry/security_authorization.go` | create | `AuthorizationContext` — empty (no perm service) |
| `internal/registry/security_http.go` | create | `HttpSecurityContext` — passthrough middleware stubs |
| `internal/registry/security_context.go` | create | `SecurityContext` — composes the three above |
| `internal/registry/service_context.go` | create | `ServiceContext` + `IngestContext` |
| `internal/registry/consumer_context.go` | create | `ConsumerContext` |
| `internal/registry/cron_context.go` | create | `CronContext` |
| `internal/service/service_factory.go` | create | `ServiceFactoryContext` + stub `ServiceFactory` |
| `cmd/api/main.go` | modify | wire `registry.NewServiceContext(c)` |
| `cmd/ingest/main.go` | modify | wire `registry.NewIngestContext(c)` |
| `cmd/worker/main.go` | modify | wire `registry.NewConsumerContext(c)` |
| `cmd/cron/main.go` | modify | wire `registry.NewCronContext(c)` |

---

## Task 1: BaseContext

**Files:**
- Create: `internal/registry/base_context.go`

- [ ] **Step 1: Write `base_context.go`**

```go
package registry

import "pulse/helper/utils/toolkit/downloader"

// BaseContext is the minimum interface embedded by all registry contexts.
type BaseContext interface {
	GetDownloader() downloader.Downloader
}

type baseContext struct{}

func newBaseContext() *baseContext { return &baseContext{} }

func (b *baseContext) GetDownloader() downloader.Downloader {
	return downloader.NewDownloader()
}
```

- [ ] **Step 2: Build to confirm no errors**

```bash
go build ./internal/registry/...
```
Expected: no output (clean build). May warn about unused package — that's fine.

---

## Task 2: DB Helper

**Files:**
- Create: `internal/registry/db.go`

- [ ] **Step 1: Write `db.go`**

```go
package registry

import (
	"pulse/internal/types/entity"

	"github.com/zeromicro/go-zero/core/logx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// mustOpenDB opens a GORM/Postgres connection and runs AutoMigrate. Panics on error.
// Called once per binary at startup — fail-fast so misconfigured services don't run.
func mustOpenDB(dsn string) *gorm.DB {
	db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	logx.Must(err)
	logx.Must(db.AutoMigrate(entity.All()...))
	return db
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```
Expected: clean.

---

## Task 3: RepositoryContext (all nil stubs)

**Files:**
- Create: `internal/registry/repository_context.go`

- [ ] **Step 1: Write `repository_context.go`**

```go
package registry

import (
	"pulse/internal/contract"

	"gorm.io/gorm"
)

// RepositoryContext exposes all persistence interfaces.
// All getters return nil until TASK-011+ wires concrete implementations.
type RepositoryContext interface {
	GetLogRepo() contract.LogRepo
	GetIncidentRepo() contract.IncidentRepo
	GetEmbeddingRepo() contract.EmbeddingRepo
	GetLLMCallRepo() contract.LLMCallRepo
	GetTenantRepo() contract.TenantRepo
	GetAPIKeyRepo() contract.APIKeyRepo
	GetOutboxRepo() contract.OutboxRepo
	GetSagaRepo() contract.SagaRepo
	GetAuditLogRepo() contract.AuditLogRepo
	GetTxManager() contract.TxManager
	GetOutboxAppender() contract.OutboxAppender
}

type repositoryContext struct {
	db *gorm.DB
}

// NewRepositoryContext creates the context from an open GORM connection.
// Replace nil returns with concrete repo constructors as each is implemented.
func NewRepositoryContext(db *gorm.DB) RepositoryContext {
	return &repositoryContext{db: db}
}

func (r *repositoryContext) GetLogRepo() contract.LogRepo              { return nil }
func (r *repositoryContext) GetIncidentRepo() contract.IncidentRepo     { return nil }
func (r *repositoryContext) GetEmbeddingRepo() contract.EmbeddingRepo   { return nil }
func (r *repositoryContext) GetLLMCallRepo() contract.LLMCallRepo       { return nil }
func (r *repositoryContext) GetTenantRepo() contract.TenantRepo         { return nil }
func (r *repositoryContext) GetAPIKeyRepo() contract.APIKeyRepo         { return nil }
func (r *repositoryContext) GetOutboxRepo() contract.OutboxRepo         { return nil }
func (r *repositoryContext) GetSagaRepo() contract.SagaRepo             { return nil }
func (r *repositoryContext) GetAuditLogRepo() contract.AuditLogRepo     { return nil }
func (r *repositoryContext) GetTxManager() contract.TxManager           { return nil }
func (r *repositoryContext) GetOutboxAppender() contract.OutboxAppender { return nil }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```
Expected: clean.

---

## Task 4: ServiceFactory stub

**Files:**
- Create: `internal/service/service_factory.go`

The real services are added in later tasks; `pmctl gen service` will regenerate this file once services exist. For now we need the type so `ServiceContext.GetServiceFactory()` compiles.

- [ ] **Step 1: Write `internal/service/service_factory.go`**

```go
package service

import (
	"pulse/helper/utils/toolkit/downloader"
	"pulse/internal/contract"
)

// ServiceFactoryContext is the subset of registry.ServiceContext / registry.CronContext
// that the ServiceFactory needs. Using this interface (not the concrete context types)
// prevents an import cycle.
type ServiceFactoryContext interface {
	GetDownloader() downloader.Downloader
	GetLogRepo() contract.LogRepo
	GetIncidentRepo() contract.IncidentRepo
	GetEmbeddingRepo() contract.EmbeddingRepo
	GetLLMCallRepo() contract.LLMCallRepo
	GetTenantRepo() contract.TenantRepo
	GetAPIKeyRepo() contract.APIKeyRepo
	GetOutboxRepo() contract.OutboxRepo
	GetSagaRepo() contract.SagaRepo
	GetAuditLogRepo() contract.AuditLogRepo
	GetTxManager() contract.TxManager
	GetOutboxAppender() contract.OutboxAppender
}

// ServiceFactory vends service instances via lazy oncex initialization.
// Run `pmctl gen service` after adding or removing a service.
type ServiceFactory struct {
	ctx ServiceFactoryContext
}

func NewServiceFactory(ctx ServiceFactoryContext) *ServiceFactory {
	return &ServiceFactory{ctx: ctx}
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/service/...
```
Expected: clean.

---

## Task 5: Security — AuthenticationContext

**Files:**
- Create: `internal/registry/security_authentication.go`

For Pulse, authentication is JWT-secret-based (not JWKS). The actual middleware lives in `internal/auth/` (wired in auth tasks). Here we just expose the secret so middleware constructors can consume it.

- [ ] **Step 1: Write `security_authentication.go`**

```go
package registry

import "pulse/internal/config"

// AuthenticationContext exposes credentials needed by JWT middleware.
type AuthenticationContext interface {
	GetJWTSecret() string
}

type authenticationContext struct {
	jwtSecret string
}

func NewAuthenticationContext(c config.APIConfig) AuthenticationContext {
	return &authenticationContext{jwtSecret: c.Auth.JwtSecret}
}

func (a *authenticationContext) GetJWTSecret() string { return a.jwtSecret }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 6: Security — AuthorizationContext

**Files:**
- Create: `internal/registry/security_authorization.go`

Pulse has no external permission service. The interface is a placeholder for future RBAC.

- [ ] **Step 1: Write `security_authorization.go`**

```go
package registry

// AuthorizationContext is a placeholder — Pulse has no external permission service.
// Extend this interface when RBAC is introduced.
type AuthorizationContext interface{}

type authorizationContext struct{}

func NewAuthorizationContext() AuthorizationContext {
	return &authorizationContext{}
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 7: Security — HttpSecurityContext

**Files:**
- Create: `internal/registry/security_http.go`

HTTP middleware stubs — pass-through for now. Real implementations replace these in the auth tasks.

- [ ] **Step 1: Write `security_http.go`**

```go
package registry

import (
	"net/http"

	"github.com/zeromicro/go-zero/rest"
)

// HttpSecurityContext provides REST middleware for the API server.
type HttpSecurityContext interface {
	// GetAuthMiddleware returns JWT authentication middleware.
	// Stub: passthrough. Replace with real JWT check in the auth task.
	GetAuthMiddleware() rest.Middleware
	// GetTenantMiddleware injects the tenant from JWT claims into context.
	// Stub: passthrough.
	GetTenantMiddleware() rest.Middleware
}

type httpSecurityContext struct {
	authMiddleware   rest.Middleware
	tenantMiddleware rest.Middleware
}

func NewHttpSecurityContext() HttpSecurityContext {
	passthrough := func(next http.HandlerFunc) http.HandlerFunc { return next }
	return &httpSecurityContext{
		authMiddleware:   passthrough,
		tenantMiddleware: passthrough,
	}
}

func (h *httpSecurityContext) GetAuthMiddleware() rest.Middleware   { return h.authMiddleware }
func (h *httpSecurityContext) GetTenantMiddleware() rest.Middleware { return h.tenantMiddleware }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 8: Security — SecurityContext composition

**Files:**
- Create: `internal/registry/security_context.go`

- [ ] **Step 1: Write `security_context.go`**

```go
package registry

import "pulse/internal/config"

// SecurityContext composes authentication, authorization, and HTTP security.
type SecurityContext interface {
	AuthenticationContext
	AuthorizationContext
	HttpSecurityContext
}

type securityContext struct {
	AuthenticationContext
	AuthorizationContext
	HttpSecurityContext
}

// NewSecurityContext wires all three security sub-contexts for the API server.
func NewSecurityContext(c config.APIConfig) SecurityContext {
	return &securityContext{
		AuthenticationContext: NewAuthenticationContext(c),
		AuthorizationContext:  NewAuthorizationContext(),
		HttpSecurityContext:   NewHttpSecurityContext(),
	}
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 9: ServiceContext + IngestContext

**Files:**
- Create: `internal/registry/service_context.go`

- [ ] **Step 1: Write `service_context.go`**

```go
package registry

import (
	"pulse/helper/utils/toolkit/oncex"
	"pulse/internal/config"
	"pulse/internal/service"
)

// ServiceContext is the root DI context for cmd/api.
// Composes all sub-contexts and vends a lazy ServiceFactory.
type ServiceContext interface {
	BaseContext
	GetConfig() config.APIConfig
	RepositoryContext
	SecurityContext
	GetServiceFactory() *service.ServiceFactory
}

type serviceContext struct {
	*baseContext
	RepositoryContext
	SecurityContext
	config         config.APIConfig
	serviceFactory oncex.OnceValue[*service.ServiceFactory]
}

// NewServiceContext opens the DB, runs AutoMigrate, and wires all sub-contexts.
// Panics if the DB or any required dependency is unavailable at startup.
func NewServiceContext(c config.APIConfig) ServiceContext {
	return &serviceContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
		SecurityContext:   NewSecurityContext(c),
	}
}

func (s *serviceContext) GetConfig() config.APIConfig { return s.config }

func (s *serviceContext) GetServiceFactory() *service.ServiceFactory {
	return s.serviceFactory.MustGet(func() *service.ServiceFactory {
		return service.NewServiceFactory(s)
	})
}

// IngestContext is the DI context for cmd/ingest.
// No DB — ingest writes to Kafka only (via the outbox pattern, TASK-032+).
type IngestContext interface {
	BaseContext
	GetConfig() config.IngestConfig
}

type ingestContext struct {
	*baseContext
	config config.IngestConfig
}

func NewIngestContext(c config.IngestConfig) IngestContext {
	return &ingestContext{
		baseContext: newBaseContext(),
		config:      c,
	}
}

func (i *ingestContext) GetConfig() config.IngestConfig { return i.config }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```
Expected: clean.

---

## Task 10: ConsumerContext

**Files:**
- Create: `internal/registry/consumer_context.go`

- [ ] **Step 1: Write `consumer_context.go`**

```go
package registry

import "pulse/internal/config"

// ConsumerContext is the DI context for cmd/worker (Kafka consumer).
type ConsumerContext interface {
	BaseContext
	GetConfig() config.WorkerConfig
	RepositoryContext
}

type consumerContext struct {
	*baseContext
	RepositoryContext
	config config.WorkerConfig
}

// NewConsumerContext opens the DB so consumers can read/write the outbox and repos.
func NewConsumerContext(c config.WorkerConfig) ConsumerContext {
	return &consumerContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
	}
}

func (c *consumerContext) GetConfig() config.WorkerConfig { return c.config }
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 11: CronContext

**Files:**
- Create: `internal/registry/cron_context.go`

- [ ] **Step 1: Write `cron_context.go`**

```go
package registry

import (
	"pulse/helper/utils/toolkit/oncex"
	"pulse/internal/config"
	"pulse/internal/service"
)

// CronContext is the DI context for cmd/cron (scheduled jobs).
type CronContext interface {
	BaseContext
	GetConfig() config.CronConfig
	RepositoryContext
	GetServiceFactory() *service.ServiceFactory
}

type cronContext struct {
	*baseContext
	RepositoryContext
	config         config.CronConfig
	serviceFactory oncex.OnceValue[*service.ServiceFactory]
}

// NewCronContext opens the DB for cron job reads/writes.
func NewCronContext(c config.CronConfig) CronContext {
	return &cronContext{
		baseContext:       newBaseContext(),
		config:            c,
		RepositoryContext: NewRepositoryContext(mustOpenDB(c.DB.DataSource)),
	}
}

func (c *cronContext) GetConfig() config.CronConfig { return c.config }

func (c *cronContext) GetServiceFactory() *service.ServiceFactory {
	return c.serviceFactory.MustGet(func() *service.ServiceFactory {
		return service.NewServiceFactory(c)
	})
}
```

- [ ] **Step 2: Build**

```bash
go build ./internal/registry/...
```

---

## Task 12: Wire cmd/ entry points

**Files:**
- Modify: `cmd/api/main.go`
- Modify: `cmd/ingest/main.go`
- Modify: `cmd/worker/main.go`
- Modify: `cmd/cron/main.go`

- [ ] **Step 1: Update `cmd/api/main.go`**

Replace the TODO comment with:

```go
package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
	"pulse/internal/config"
	"pulse/internal/handler"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/api.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.APIConfig](configFile)
	fmt.Printf("Pulse API server starting (name: %s, http: %s:%d, grpc: %s)\n",
		c.Name, c.Host, c.Port, c.Grpc.ListenOn)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	svcGroup.Add(handler.NewHealthServer(c.Name, c.Host, c.Port))

	_ = registry.NewServiceContext(c) // TODO TASK-023: wire HTTP + gRPC handlers

	svcGroup.Start()
}
```

- [ ] **Step 2: Update `cmd/ingest/main.go`**

```go
package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
	"pulse/internal/config"
	"pulse/internal/handler"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/ingest.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.IngestConfig](configFile)
	fmt.Printf("Pulse Ingest server starting (name: %s, http: %s:%d)\n",
		c.Name, c.Host, c.Port)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	svcGroup.Add(handler.NewHealthServer(c.Name, c.Host, c.Port))

	_ = registry.NewIngestContext(c) // TODO TASK-032: wire ingestion HTTP server

	svcGroup.Start()
}
```

- [ ] **Step 3: Update `cmd/worker/main.go`**

```go
package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"github.com/zeromicro/go-zero/core/service"
	"pulse/internal/config"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/worker.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.WorkerConfig](configFile)
	fmt.Printf("Pulse Worker starting (name: %s, brokers: %v)\n",
		c.Name, c.Kafka.Brokers)

	svcGroup := service.NewServiceGroup()
	defer svcGroup.Stop()

	_ = registry.NewConsumerContext(c) // TODO TASK-016: wire Kafka consumers + outbox drainer

	svcGroup.Start()
}
```

- [ ] **Step 4: Update `cmd/cron/main.go`**

```go
package main

import (
	"flag"
	"fmt"

	"github.com/joho/godotenv"
	"pulse/internal/config"
	"pulse/internal/registry"
)

var configFile = flag.String("f", "etc/cron.yaml", "the config file")

func main() {
	flag.Parse()
	_ = godotenv.Load()

	c := config.Load[config.CronConfig](configFile)
	fmt.Printf("Pulse Cron starting (name: %s, budget: $%.2f/day)\n",
		c.Name, c.LLM.DailyBudgetUsd)

	_ = registry.NewCronContext(c) // TODO TASK-038: register cron commands
}
```

---

## Task 13: Verify full build

- [ ] **Step 1: Build all four binaries**

```bash
go build ./cmd/api/... ./cmd/ingest/... ./cmd/worker/... ./cmd/cron/...
```
Expected: no output (clean).

- [ ] **Step 2: Run vet**

```bash
go vet ./internal/registry/... ./internal/service/... ./cmd/...
```
Expected: no warnings.

- [ ] **Step 3: Commit**

```bash
git add internal/registry/ internal/service/service_factory.go cmd/api/main.go cmd/ingest/main.go cmd/worker/main.go cmd/cron/main.go
git commit -m "feat(registry): implement DI composition root for all four binaries"
```

---

## Verification

```bash
# Compile check — no DB needed
go build ./cmd/api/... ./cmd/ingest/... ./cmd/worker/... ./cmd/cron/...

# Smoke test health endpoints (DB not required — health server wired before svcCtx)
go build -o /tmp/pulse-api ./cmd/api && go build -o /tmp/pulse-ingest ./cmd/ingest

SERVICE_API_HTTP_PORT=8000 SERVICE_API_GRPC_PORT=8001 \
  DB_HOST=localhost DB_PORT=5433 DB_NAME=pulse_db DB_USER=postgres DB_PASSWORD=postgres \
  REDIS_HOST=localhost REDIS_PORT=6379 KAFKA_BROKERS=localhost:9092 \
  JWT_SECRET=test ANTHROPIC_API_KEY=test ANTHROPIC_MODEL=claude-sonnet-4-6 \
  LLM_DAILY_BUDGET_USD=5.0 TELEMETRY_ENDPOINT=http://localhost:4318 \
  /tmp/pulse-api -f etc/api.yaml &
sleep 2 && curl -s localhost:8000/health
# → {"status":"ok"}

# API and Ingest servers panic at startup if DB is unreachable — that's expected and correct.
# Full integration test requires docker-compose up (TimescaleDB, Redis, Kafka).
```

---

## Notes

- `ServiceFactory` and `CronContext.GetServiceFactory()` both accept `ServiceFactoryContext` which requires only `GetDownloader()` + all repo getters — NOT a typed `GetConfig()`. This avoids a cross-config-type interface mismatch.
- All repo getters return `nil`. Any call to a repo method before TASK-011+ wires it will panic at runtime — this is intentional and correct (fail fast, fail loudly).
- Security middleware is passthrough. No JWT validation happens until the auth task replaces `NewHttpSecurityContext()` internals.
- `IngestContext` has no DB because ingest publishes to Kafka via the Outbox pattern; the outbox is owned by the API/worker side.
