# Phase 2 Summary: Reference Domain — `auth`

## Overview

Phase 2 implements the `auth` domain as the **reference vertical slice** for all other Pulse domains. It covers TASK-020 through TASK-024 and establishes the pattern: repository → service → middleware → handler → registry wiring.

## What Already Exists

| Layer | Status |
|---|---|
| **Entities** | Tenant, User, APIKey, AuditLog — fully defined in `internal/types/entity/` |
| **Contracts** | TenantRepo, APIKeyRepo, AuditLogRepo interfaces defined in `internal/contract/repository.go` (UserRepo missing) |
| **DB Migrations** | Tables `tenants`, `users`, `api_keys`, `audit_logs` created via `db/migrations/0001_init.up.sql` |
| **DI Container** | RepositoryContext getters return `nil`; SecurityContext middleware are passthroughs |
| **Auth Stubs** | 8 files in `internal/auth/` — all empty (`package auth` only) |
| **Seed Data** | `scripts/seed.sh` — demo tenant (slug=demo), user (demo@pulse.dev/demo1234), API key (pk_demo_xxx) |
| **Helper Packages** | JWT parsing (`identity/parser.go`), claims model (`identity/claims.go`), error types (`errors/`), HTTP response (`response/`) |

## Design Decisions

1. **JWT HS256** — HMAC-SHA256 with `AuthConfig.JwtSecret`. Compatible with existing `identity.FromToken()` parser.
2. **Hand-written REST** — No proto for auth. Routes registered directly on go-zero `rest.Server`.
3. **Stateless refresh** — Access (15 min) + refresh (30 day) JWT. No server-side token storage.
4. **Claims structure** — `eid` (user ID), `kind` (user/refresh), `attributes.tenant_id`, `attributes.email` — compatible with `identity.NewClaimsFromJWT()`.
5. **Response envelope** — `{"meta": {...}, "data": {...}}` via `response.OkJson()`.

## Task Breakdown

### TASK-020: Repositories
- Add `UserRepo` interface to `internal/contract/repository.go`
- Add `GetUserRepo()` to `RepositoryContext` and `ServiceFactoryContext`
- Implement 4 repos in `internal/auth/`:
  - `tenant_repo.go` — TenantRepo (Save, FindByID, FindBySlug, Update)
  - `user_repo.go` — UserRepo (FindByEmail, FindByID) *new file*
  - `api_key_repo.go` — APIKeyRepo (Save, FindByHash, ListByTenant, Delete, TouchLastUsed)
  - `audit_repo.go` — AuditLogRepo (Append, FindByResource)

### TASK-021: AuthService
- Create `internal/auth/dto.go` — LoginRequest, RefreshRequest, CreateAPIKeyRequest, TokenResponse, APIKeyResponse, etc.
- Implement `internal/auth/service.go`:
  - `Login()` — email/password → bcrypt → JWT pair → audit
  - `Refresh()` — parse refresh JWT → verify kind → new token pair
  - `CreateAPIKey()` — generate `pk_<hex>` → SHA-256 hash → save → return raw key once
  - `RevokeAPIKey()` — soft-delete → audit
  - `ListAPIKeys()` — list by tenant (never expose hash)

### TASK-022: Middleware
- `auth_middleware.go` — JWT Bearer verification, rejects refresh tokens, injects claims into context
- `tenant_middleware.go` — Extracts `tenant_id` from claims → context. Exports `TenantIDFromContext()`
- `hmac_middleware.go` — Verifies `X-Signature: sha256=<hex>` HMAC of request body

### TASK-023: Handler + Wiring
- Implement `internal/auth/handler.go` — 5 endpoints:
  - `POST /v1/auth/login` (public)
  - `POST /v1/auth/refresh` (public)
  - `POST /v1/auth/keys` (protected)
  - `DELETE /v1/auth/keys/:id` (protected)
  - `GET /v1/auth/keys` (protected)
- Wire repos in `registry/repository_context.go` (replace nil → concrete)
- Wire real middleware in `registry/security_http.go` (replace passthrough → JWT + tenant)
- Create `RestHandler.Register()` in `handler/http_handler.go`
- Update `cmd/api/main.go` to use ServiceContext + RestHandler

### TASK-024: Smoke Test
- `scripts/smoke-auth.sh` — 6-step E2E: login → refresh → create key → list keys → revoke key → verify 401

## Files Changed (17 files)

| File | Action |
|---|---|
| `internal/contract/repository.go` | Add UserRepo interface |
| `internal/registry/repository_context.go` | Wire 4 auth repos |
| `internal/service/service_factory.go` | Add GetUserRepo() |
| `internal/auth/tenant_repo.go` | Implement TenantRepository |
| `internal/auth/user_repo.go` | **New** — UserRepository |
| `internal/auth/api_key_repo.go` | Implement APIKeyRepository |
| `internal/auth/audit_repo.go` | Implement AuditLogRepository |
| `internal/auth/dto.go` | **New** — DTOs |
| `internal/auth/service.go` | Implement AuthService |
| `internal/auth/auth_middleware.go` | JWT middleware |
| `internal/auth/tenant_middleware.go` | Tenant middleware |
| `internal/auth/hmac_middleware.go` | HMAC middleware |
| `internal/auth/handler.go` | REST handler (5 endpoints) |
| `internal/handler/http_handler.go` | RestHandler wiring |
| `internal/registry/security_http.go` | Real middleware |
| `cmd/api/main.go` | Wire everything |
| `scripts/smoke-auth.sh` | **New** — Smoke test |

## Key Dependencies & Reuse

| Reusable Code | Location |
|---|---|
| JWT parsing | `helper/utils/identity/parser.go` → `FromToken()` |
| Claims model | `helper/utils/identity/claims.go` → `MapClaims` |
| Claims↔context | `helper/utils/identity/context.go` → `WithContext()`, `FromContext()` |
| Error types | `helper/utils/errors/` → `NewUnauthorized()`, `NewBadRequest()` |
| HTTP response | `helper/utils/server/http/response/` → `OkJson()`, `Error()` |
| Repo pattern | `internal/outbox/repository.go` → GORM + WithContext + error wrapping |
| bcrypt | `golang.org/x/crypto/bcrypt` (in go.mod) |
| JWT lib | `github.com/golang-jwt/jwt/v4` (in go.mod) |
| UUID | `github.com/google/uuid` (in go.mod) |

## Verification

1. `go build ./...` — compiles clean
2. `docker compose up -d` → `make seed` → `make run`
3. `bash scripts/smoke-auth.sh` — all 6 steps pass
4. Protected endpoints return 401 without token
5. Refresh token rejected when used as access token
6. API key creation returns raw key; list never exposes hash

## Exit Criteria

`bash scripts/smoke-auth.sh` passes end-to-end. Auth domain becomes the template — read `internal/auth/` before writing any other domain.
