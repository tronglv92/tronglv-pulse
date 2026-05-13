---
description: Execute a Go backend implementation plan — build feature, validate, and finalize
name: /execute
argument-hint: [feature-name-or-plan]
---

# Go Backend Service — Execution Workflow

## MCP Tool Usage Rules

Apply these rules throughout every execution session:

### Skills
- Always check available skills before starting — load `systematic-debugging` for bugs, `test-driven-development` for features
- Never skip skill loading to save time — skills prevent rework

### GoAtlas (Code Intelligence)
When you need to explore source code, check structure, find functions/types/endpoints, or understand code flow, prefer using MCP GoAtlas tools (`search_code`, `find_symbol`, `find_callers`, `get_file_symbols`, `list_api_endpoints`, `read_file`) over manual grep/find. GoAtlas has indexed all PMC WMS services.

### Token Budget — Task Classification

Classify each task before choosing a review strategy. Wrong choice wastes 10–30× tokens.

| Task type | Examples | Review strategy |
|---|---|---|
| **Mechanical** | Interface files, type/constant definitions, config structs, proto stub writes | 1 implementer subagent → `go build` + `go vet` → commit. No spec/quality reviewers. |
| **Standard** | CRUD repository, simple service method, mapper, handler wiring | 1 implementer → spec reviewer → commit |
| **Complex** | Business logic with edge cases, auth, multi-service coordination, anomaly/RCA pipeline | Full pipeline: implementer → spec reviewer → code quality reviewer |

**For exploration:** Always call GoAtlas MCP tools first (`find_symbol`, `get_file_symbols`). Only spawn an Explore subagent if GoAtlas cannot answer the question. Never spawn 3 Explore agents when 1 GoAtlas call answers it.

**For planning:** Skip the Plan subagent when the task spec already defines exact file contents (interfaces, config, protos). Write the plan directly from spec. Plan subagents are for tasks requiring design judgment.

### Technical Specs
- Read proto files and entity definitions before writing any code
- When in doubt about field names or types, check `api/{domain}/*.proto` first, then existing entity files

### Security Rules
- Never hardcode credentials, tokens, secrets, or connection strings
- All authentication via middleware; authorization via `permission-svc` gRPC call
- Validate all inputs at handler level (proto validators + `protoc-gen-validate`) and service level

### Database Design Conventions
- Integer IDs for all primary and foreign keys
- GORM column tags: snake_case — `gorm:"column:transfer_order_id"`
- Index required on every foreign key and `status` field
- Use `db.WithContext(ctx)` on every GORM call — no exceptions
- Wrap multi-step writes in a single transaction

### Golang-Specific Rules
- Define interface before concrete type — always
- Constructor pattern: `New{Domain}X(deps...) *{domain}X`
- Receiver: `h` (handler), `s` (service), `r` (repository)
- Import groups: stdlib → external packages → internal packages
- Error wrapping: `fmt.Errorf("failed to ...: %w", err)` — never swallow errors
- No global variables — use dependency injection

---

## Step 0: Plan Activation

1. Load the plan (from `/plan` output or user instructions)
2. Read `CLAUDE.md` for project overrides
3. Confirm target domain files exist; if any are missing, create them in dependency order
4. Check git branch is correct before writing any file

---

## Step 1: Pre-Implementation Checks

Before writing code:

- [ ] Proto file exists and is correct (`api/{domain}/{domain}.proto`)
- [ ] Entity struct reviewed (`internal/types/entity/{domain}.go`)
- [ ] Repository interface reviewed (`internal/repository/{domain}_repository.go`)
- [ ] Registry wiring locations identified (`internal/registry/service_context.go`, `repository_context.go`)
- [ ] No conflicting implementation already exists

---

## Step 2: Implementation Order (MANDATORY — follow strictly)

### 2.1 Proto → Generated Code (if new endpoint)

```bash
# Edit proto first
vi api/{domain}/{domain}.proto

# Regenerate all proto code
make all
```

Verify generated files exist:
- `api/{domain}/{domain}.pb.go`
- `api/{domain}/{domain}_http.pb.go` (REST gateway)
- `api/{domain}/{domain}_grpc.pb.go` (gRPC)

### 2.2 Entity (if new table or columns needed)

```go
// internal/types/entity/{domain}.go
package entity

type {Domain} struct {
    gorm.Model                                              // id, created_at, updated_at, deleted_at
    TransferOrderID int    `gorm:"column:transfer_order_id;index"`
    Status          int    `gorm:"column:status;index"`
    // ... other fields
}
```

Rules:
- Embed `gorm.Model` (gives soft delete for free)
- Column tags must be snake_case
- Index tag on every FK and status field
- No business logic in entity structs

### 2.3 Repository Interface + Implementation

```go
// internal/repository/{domain}_repository.go
package repository

// Interface FIRST
type {Domain}Repository interface {
    Create(ctx context.Context, m *entity.{Domain}) (*entity.{Domain}, error)
    GetByID(ctx context.Context, id int) (*entity.{Domain}, error)
    Update(ctx context.Context, m *entity.{Domain}) error
    Delete(ctx context.Context, id int) error
    List(ctx context.Context, filter ...) ([]*entity.{Domain}, error)
}

// Implementation
type {domain}Repository struct {
    db *gorm.DB
}

func New{Domain}Repository(db *gorm.DB) {Domain}Repository {
    return &{domain}Repository{db: db}
}

func (r *{domain}Repository) Create(ctx context.Context, m *entity.{Domain}) (*entity.{Domain}, error) {
    if err := r.db.WithContext(ctx).Create(m).Error; err != nil {
        return nil, fmt.Errorf("failed to create {domain}: %w", err)
    }
    return m, nil
}
```

Register in `internal/registry/repository_context.go`:
```go
ctx.{Domain}Repository = repository.New{Domain}Repository(ctx.Db)
```

### 2.4 Service Interface + Implementation

```go
// internal/service/{domain}_service.go
package service

// Interface FIRST
type {Domain}Service interface {
    Create{Domain}(ctx context.Context, req *Request) (*Response, error)
    Get{Domain}(ctx context.Context, id int) (*entity.{Domain}, error)
}

// Implementation
type {domain}Service struct {
    repo  repository.{Domain}Repository
    cache cache.Cache          // only if idempotency needed
}

func New{Domain}Service(repo repository.{Domain}Repository) {Domain}Service {
    return &{domain}Service{repo: repo}
}
```

Register in `internal/registry/service_context.go`:
```go
ctx.{Domain}Service = service.New{Domain}Service(ctx.{Domain}Repository)
```

### 2.5 Handler

```go
// internal/handler/{domain}_handler.go
package handler

type {domain}Handler struct {
    ctx     context.Context
    service service.{Domain}Service
}

func New{Domain}Handler(
    ctx context.Context,
    svc service.{Domain}Service,
) *{domain}Handler {
    return &{domain}Handler{ctx: ctx, service: svc}
}

// gRPC method
func (h *{domain}Handler) Create{Domain}(
    ctx context.Context,
    req *{domain}proto.Create{Domain}Request,
) (*{domain}proto.Create{Domain}Response, error) {
    result, err := h.service.Create{Domain}(ctx, mapper.MapRequestTo{Domain}(req))
    if err != nil {
        return nil, status.Error(codes.Internal, err.Error())
    }
    return &{domain}proto.Create{Domain}Response{
        Data: mapper.Map{Domain}ToProto(result),
    }, nil
}
```

Register gRPC handler in `internal/handler/grpc_handler.go`:
```go
{domain}proto.Register{Domain}ServiceServer(server, handler.New{Domain}Handler(ctx, ctx.{Domain}Service))
```

Register REST handler in `internal/handler/http_handler.go` (if REST endpoint):
```go
{domain}proto.Register{Domain}ServiceHandlerFromEndpoint(...)
```

### 2.6 Mapper

```go
// internal/types/mapper/{domain}_mapper.go
package mapper

func Map{Domain}ToProto(e *entity.{Domain}) *{domain}proto.{Domain} {
    if e == nil {
        return nil
    }
    return &{domain}proto.{Domain}{
        Id:              int32(e.ID),
        TransferOrderId: int32(e.TransferOrderID),
        Status:          int32(e.Status),
        CreatedAt:       timestamppb.New(e.CreatedAt),
    }
}

func MapRequestTo{Domain}(req *{domain}proto.Create{Domain}Request) *entity.{Domain} {
    return &entity.{Domain}{
        TransferOrderID: int(req.TransferOrderId),
        Status:          constant.{Domain}StatusPending,
    }
}
```

### 2.7 Constants (if new status values)

```go
// internal/types/define/constant/{domain}.go
package constant

const (
    {Domain}StatusPending   = 1
    {Domain}StatusActive    = 2
    {Domain}StatusCompleted = 3
    {Domain}StatusCancelled = 4
)
```

### 2.8 Consumer / Cron (if async processing)

- Consumer: `internal/consumer/consumer.go` — register new topic handler
- Cron: `internal/cron/{domain}_cronjob.go` + register in `internal/cron/command.go`

---

## Step 3: Code Quality Rules

### Error Handling (MANDATORY)

```go
// Handler — map errors to gRPC status codes
if errors.Is(err, service.Err{Domain}NotFound) {
    return nil, status.Error(codes.NotFound, err.Error())
}
if errors.Is(err, service.ErrInvalid{Domain}) {
    return nil, status.Error(codes.InvalidArgument, err.Error())
}
return nil, status.Error(codes.Internal, err.Error())

// Service — wrap with context
return nil, fmt.Errorf("{domain}: %w", err)

// Repository — wrap with table context
return nil, fmt.Errorf("failed to create {domain}: %w", err)
```

### Idempotency Pattern (for create/allocate operations)

```go
cacheKey := fmt.Sprintf("{domain}:create:%d", req.TransferOrderID)
if cached, err := s.cache.Get(ctx, cacheKey); err == nil {
    return cached.(*entity.{Domain}), nil
}
// ... create ...
s.cache.Set(ctx, cacheKey, result, 30*time.Minute)
```

### Transactions (for multi-step writes)

```go
err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
    if err := tx.Create(parent).Error; err != nil {
        return err
    }
    for _, child := range children {
        child.ParentID = parent.ID
        if err := tx.Create(child).Error; err != nil {
            return err
        }
    }
    return nil
})
```

### Logging

```go
logger := log.GetLogger()
logger.Info("{domain} created",
    zap.Int("{domain}_id", result.ID),
    zap.Int("transfer_order_id", req.TransferOrderID),
)
logger.Error("failed to create {domain}",
    zap.Error(err),
    zap.Int("transfer_order_id", req.TransferOrderID),
)
```

---

## Step 4: Validation

### 4.1 Build

```bash
go build ./...
```

Fix all compilation errors before proceeding.

### 4.2 Lint

```bash
golangci-lint run ./...
# Auto-fix
gofmt -s -w .
```

### 4.3 Tests

```bash
go test ./...
go test -cover ./...
```

### 4.4 Manual Checklist

- [ ] `go build ./...` passes with zero errors
- [ ] `golangci-lint run ./...` passes
- [ ] New repository registered in `internal/registry/repository_context.go`
- [ ] New service registered in `internal/registry/service_context.go`
- [ ] New handler registered in `internal/handler/grpc_handler.go` (and `http_handler.go` if REST)
- [ ] Interface defined before implementation
- [ ] All GORM calls use `db.WithContext(ctx)`
- [ ] No hardcoded credentials or magic numbers
- [ ] Errors wrapped with `fmt.Errorf("...: %w", err)`
- [ ] Proto regenerated if proto was modified (`make all`)
- [ ] Constants in `internal/types/define/` (no inline magic values)
- [ ] Mapper covers all proto ↔ entity fields (nil-safe)

---

## Step 5: Summary

```
=== Execution Complete ===

Feature:  {name}
Domain:   {domain}

Files created:
  - {path} — {purpose}

Files modified:
  - {path} — {what changed}

Registry wired:
  - {domain}Repository → repository_context.go
  - {domain}Service    → service_context.go
  - {domain}Handler    → grpc_handler.go

Validation:
  - go build:        [pass/fail]
  - golangci-lint:   [pass/fail]
  - go test:         [pass/fail]

Next steps:
  - {any remaining work}
```

---

## Execution Rules

1. **Interface first** — always define the interface before writing the implementation
2. **Dependency order** — entity → repository → service → handler → mapper → registry
3. **Proto first** — if the feature touches the API, regenerate protos before writing Go code
4. **WithContext always** — every GORM call must have `db.WithContext(ctx)`
5. **Registry wiring** — every new component must be registered in the appropriate registry file
6. **No hardcoded values** — use constants from `internal/types/define/`
7. **Wrap errors** — never return raw errors; always add context with `fmt.Errorf`
8. **Minimal changes** — only write files the plan requires; no unsolicited refactoring
