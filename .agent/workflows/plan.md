---
description: Plan the implementation of a feature, API endpoint, or fix — design the approach before writing any Go code
---

# Go Backend Service — Planning Workflow

## Purpose

Design a complete, ordered implementation plan for any feature or fix.
This workflow produces a task list — **no code is written** during planning.

---

## MCP Tool Usage Rules

Apply these rules throughout every planning session:

### Skills
- Always check available skills before starting any task
- Run `Skill` tool to load relevant skills (`systematic-debugging`, `writing-plans`, etc.)
- Skills update automatically — always load the current version, never rely on memory

### NotebookLM (Spec & Docs)
When you need to check specifications, business documentation, or need context about business logic, proactively use MCP NotebookLM tools (`notebook_query`, `notebook_list`) to look up information from existing notebooks. Especially when implementing new features or fixing bugs related to business logic, always cross-check with specs on NotebookLM before writing code.

### Technical Specs
- When clarifying specs, business requirements, or domain data: read existing proto files, service code, and entity definitions first
- Do not assume API contracts — read `api/{domain}/*.proto` and `internal/types/` before planning changes

### Security Rules
- Never plan hardcoded credentials, tokens, or secrets in code
- Authentication must go through middleware; authorization via `permission-svc`
- All inputs validated at handler level (proto validators) and service level (business rules)

### Database Design Conventions
- Always use integer/numeric IDs for primary and foreign keys (not strings/UUIDs unless required)
- Use `gorm.Model` base struct (soft deletes via `deleted_at`)
- Follow snake_case for column names via GORM tags; indexes required on foreign keys and status fields
- Use transactions for multi-step write operations

### Golang-Specific Rules
- Depend on interfaces, not concrete types — define interface before implementation
- Constructor injection: `New{Domain}Service(repo {Domain}Repository) *{domain}Service`
- Receiver naming: `h` for handlers, `s` for services, `r` for repositories
- Group imports: stdlib → external → internal (no mixed groups)
- Always use `db.WithContext(ctx)` — never call GORM without context

### Reference Implementation
Before planning any feature or fix, always read `/Users/luongvantrong/pharmacity/backend/tronglv-log-analysic/tronglv-pulse` first to understand existing implementations, patterns, and conventions used in the project.

### Helper / Shared Utilities
When working on anything related to helpers, shared utilities, or common logic used across services, always check `/Users/luongvantrong/pharmacity/backend/tronglv-log-analysic/helper` first to understand existing implementations and avoid duplication.

### Unrelated API Protection
Never plan or make changes to the response structure, fields, or behavior of APIs not directly related to the current task. If such a change seems necessary, stop and ask the user for confirmation before proceeding. Wait for explicit approval.

---

## Step 1: Load Context

1. Run the `/init` workflow if context is not already loaded
2. Read `CLAUDE.md` for project-specific overrides
3. Read `/Users/luongvantrong/pharmacity/backend/tronglv-log-analysic/tronglv-pulse` to understand existing implementations, patterns, and conventions before planning anything
4. Confirm the target domain exists in `api/`, `internal/handler/`, `internal/service/`, `internal/repository/`

---

## Step 2: Identify Domain & Scope

For the task, determine:

| Item | How to find |
|------|-------------|
| Domain name | From task description — maps to `{domain}` across all layers |
| Proto file | `api/{domain}/{domain}.proto` |
| Handler | `internal/handler/{domain}_handler.go` |
| Service | `internal/service/{domain}_service.go` |
| Repository | `internal/repository/{domain}_repository.go` |
| Entity | `internal/types/entity/{domain}.go` |
| DTO | `internal/types/dto/{domain}.go` |
| Mapper | `internal/types/mapper/{domain}_mapper.go` |
| Constants | `internal/types/define/constant/` or `define/enum/` |

If the domain is **new**, plan creation of all these files.
If **existing**, read the current files before planning changes.

---

## Step 3: API / Proto Verification (MANDATORY before designing changes)

When the task involves adding or modifying an API endpoint, **verify before planning code**.

### 3.1 Read the Proto Contract

```bash
cat api/{domain}/{domain}.proto
```

Identify:
- Service methods: `rpc MethodName(Request) returns (Response)`
- Request/response message fields and validation rules
- HTTP gateway annotations (if REST endpoint)

### 3.2 Check Existing Endpoint (if modifying)

Test the existing gRPC or REST endpoint to understand current behavior:

```bash
# REST via HTTP gateway
curl -s -H "Authorization: Bearer <TOKEN>" \
  "http://localhost:8080/api/v1/{endpoint}" | jq '.'

# gRPC via grpcurl
grpcurl -plaintext \
  -H "Authorization: Bearer <TOKEN>" \
  -d '{"field": "value"}' \
  localhost:9090 \
  pmc.{domain}.v1.{Domain}Service/{Method}
```

### 3.3 Verify Response Structure

From the actual response, confirm:

| Field to verify | Where to check |
|----------------|----------------|
| Response wrapper | Proto message definition |
| Repeated fields | `repeated` keyword in proto |
| Status/enum values | `internal/types/define/` constants |
| Error codes | gRPC status codes used in handler |

### 3.4 Map Proto Fields → Go Entity

For new endpoints, plan the mapping table:

| Proto field | Go entity field | Type | Notes |
|------------|----------------|------|-------|
| `transfer_order_id` | `TransferOrderID` | `int32` → `int` | GORM column: `transfer_order_id` |
| `status` | `Status` | `int32` | Use constant from `define/` |
| `created_at` | `CreatedAt` | `*timestamppb.Timestamp` → `time.Time` | UTC |

### 3.5 Checklist before Step 4

- [ ] Proto file read and understood
- [ ] Existing behavior verified (if modifying)
- [ ] Proto ↔ entity field mapping planned
- [ ] Status constants identified in `internal/types/define/`
- [ ] Error codes planned (gRPC status codes)

---

## Step 4: Design Feature Structure

Plan files to create or modify for each layer, in dependency order:

### Layer 1 — Entity (if new DB table or columns)
```
internal/types/entity/{domain}.go
```
- GORM struct with all fields, tags, and indexes
- Embed `gorm.Model` unless custom primary key needed

### Layer 2 — Repository (data access)
```
internal/repository/{domain}_repository.go
```
- Define interface first: `Create`, `GetByID`, `Update`, `Delete`, `List...`
- Implement with GORM using `db.WithContext(ctx)`
- Register in `internal/registry/repository_context.go`

### Layer 3 — Service (business logic)
```
internal/service/{domain}_service.go
```
- Define interface first with all public methods
- Implement with idempotency if needed (Redis cache key)
- Use repository interface (never concrete type)
- Register in `internal/registry/service_context.go`

### Layer 4 — Handler (transport layer)
```
internal/handler/{domain}_handler.go
```
- Parse proto request → validate → call service → return proto response
- gRPC method: register in `internal/handler/grpc_handler.go`
- REST endpoint: register in `internal/handler/http_handler.go`

### Layer 5 — DTO & Mapper
```
internal/types/dto/{domain}.go          (if REST response ≠ proto)
internal/types/mapper/{domain}_mapper.go
```
- `Map{Domain}ToDTO(e *entity.{Domain}) *dto.{Domain}Response`
- `MapDTOTo{Domain}(dto *...) *entity.{Domain}`

### Layer 6 — Proto (if new endpoint)
```
api/{domain}/{domain}.proto
```
- Define new rpc method
- Add request/response messages with validation rules
- Regenerate: `make all`

### Layer 7 — Constants (if new status values)
```
internal/types/define/constant/{domain}.go
```

---

## Step 5: Implementation Task List

Produce an ordered task list using this format:

```
### Task {N}: {Title}
File:    {exact path to create or modify}
Pattern: {reference file from existing codebase}
Action:  {create | modify | generate}
Notes:   {any gotchas or dependencies}
```

**Standard task order:**
1. Add/update proto definitions → regenerate code (`make all`)
2. Create/update entity struct
3. Create/update repository interface + implementation
4. Create/update service interface + implementation
5. Create/update handler
6. Create/update mapper
7. Register in registry context
8. Add constants if needed
9. Update consumer/cron if async processing needed

---

## Step 6: Risk Assessment

Identify potential issues:

- **Breaking API changes**: Does modifying the proto change existing callers?
- **DB migration**: Does the entity change require a schema migration?
- **Concurrency**: Do concurrent requests need pessimistic or optimistic locking?
- **Idempotency**: Can the operation be safely retried?
- **External service dependency**: Does this call a gRPC service that could fail?
- **Kafka events**: Does this operation need to publish an event?

---

## Step 7: Plan Output

```markdown
# Plan: {Feature Name}

## Domain: {domain name}
## Scope: {new feature | modify existing | bug fix}

## Files to Create
- {path} — {purpose}

## Files to Modify
- {path} — {what changes and why}

## Tasks (ordered)
1. {task 1}
2. {task 2}
...

## API Contract
- Method: {rpc or REST}
- Request: {fields}
- Response: {fields}
- Error codes: {gRPC status codes}

## Risks
- {risk and mitigation}
```

**Next step:** Review and approve, then run `/execute`.

---

## Constraints

- **No code** — produce plan structure and task list only
- **Interface first** — always plan interface definition before implementation
- **Read before planning** — always read existing files before planning changes to them
- **Minimal scope** — plan only what is needed for the task, nothing more
- **Registry wiring** — every new repository/service must be planned for registration in `internal/registry/`
- **No unrelated API changes** — never plan changes to APIs outside the current task scope without user confirmation
