# Phase 2: Reference Domain — `auth`

## Context

The `auth` domain is the **reference implementation** for all other Pulse domains. It establishes the end-to-end vertical slice pattern: repository → service → middleware → handler → registry wiring. Every subsequent domain (search, ingestion, incident, etc.) follows this template. Currently, stub files exist in `internal/auth/` (all empty `package auth`) and the DI container returns `nil` for all auth-related repositories.

**Goal:** Complete `internal/auth/` so that `POST /v1/auth/login` returns JWT tokens, API key CRUD works, and middleware protects routes.

---

## Design Decisions

1. **JWT HS256 signing** — Use `jwt.SigningMethodHS256` with `AuthConfig.JwtSecret`. The existing `identity.FromToken()` in `helper/utils/identity/parser.go` already handles HMAC keys. No RSA infrastructure needed.

2. **Hand-written REST handlers** — No `api/auth/auth.proto` exists. Auth endpoints are REST-only (no gRPC use case). Register routes directly on the go-zero `rest.Server` using `srv.AddRoutes()`, consistent with `handler/health_handler.go`.

3. **Response envelope** — Use `response.OkJson()` / `response.Error()` from `helper/utils/server/http/response/`. Responses wrap in `{"meta": {...}, "data": ...}` format.

4. **JWT claims structure** — Use `eid` (user ID), `kind` ("user" vs "refresh"), `attributes.tenant_id`, `attributes.email` — compatible with `identity.NewClaimsFromJWT()` parser.

5. **Stateless refresh rotation** — Both access (15 min) and refresh (30 day) are JWT. No server-side token storage. Refresh token uses `kind: "refresh"` to prevent use as access token.

6. **UserRepo contract needed** — `contract/repository.go` lacks `UserRepo`. Must add before implementation.

---

## Implementation Plan

### TASK-020: Repositories

**Step 1.** Add `UserRepo` interface to `internal/contract/repository.go`
```go
type UserRepo interface {
    FindByEmail(ctx context.Context, email string) (*entity.User, error)
    FindByID(ctx context.Context, id int64) (*entity.User, error)
}
```

**Step 2.** Add `GetUserRepo() contract.UserRepo` to:
- `internal/registry/repository_context.go` — interface + nil stub
- `internal/service/service_factory.go` — `ServiceFactoryContext` interface

**Step 3.** Implement repos (all follow `internal/outbox/repository.go` pattern):

| File | Interface | Methods |
|---|---|---|
| `internal/auth/tenant_repo.go` | `contract.TenantRepo` | Save, FindByID, FindBySlug, Update |
| `internal/auth/user_repo.go` *(new)* | `contract.UserRepo` | FindByEmail, FindByID |
| `internal/auth/api_key_repo.go` | `contract.APIKeyRepo` | Save, FindByHash, ListByTenant, Delete, TouchLastUsed |
| `internal/auth/audit_repo.go` | `contract.AuditLogRepo` | Append, FindByResource |

Pattern for each:
- `var _ contract.XRepo = (*XRepository)(nil)` compile-time check
- Constructor: `func NewXRepository(db *gorm.DB) *XRepository`
- Receiver: `r`
- `r.db.WithContext(ctx)` on every query
- Error wrapping: `fmt.Errorf("failed to ...: %w", err)`

### TASK-021: AuthService

**Step 1.** Create `internal/auth/dto.go` — request/response DTOs:
- `LoginRequest{Email, Password}`
- `RefreshRequest{RefreshToken}`
- `CreateAPIKeyRequest{Name, ExpiresAt}`
- `TokenResponse{AccessToken, RefreshToken, ExpiresIn}`
- `APIKeyResponse{ID, Name, RawKey, CreatedAt, ExpiresAt}`
- `APIKeyListResponse{Keys []APIKeyItem}`

**Step 2.** Implement `internal/auth/service.go`:

Constructor:
```go
func NewAuthService(
    userRepo   contract.UserRepo,
    tenantRepo contract.TenantRepo,
    apiKeyRepo contract.APIKeyRepo,
    auditRepo  contract.AuditLogRepo,
    jwtSecret  string,
) *AuthService
```

Methods:
- **`Login(ctx, LoginRequest) (*TokenResponse, error)`** — FindByEmail → bcrypt compare → check IsActive → generate token pair → audit log
- **`Refresh(ctx, RefreshRequest) (*TokenResponse, error)`** — parse JWT → verify kind="refresh" → find user → check active → new token pair
- **`CreateAPIKey(ctx, tenantID, actorID, CreateAPIKeyRequest) (*APIKeyResponse, error)`** — generate `pk_<48-hex>` → SHA-256 hash → save → audit → return raw key once
- **`RevokeAPIKey(ctx, id, tenantID, actorID) error`** — soft-delete → audit
- **`ListAPIKeys(ctx, tenantID) (*APIKeyListResponse, error)`** — list by tenant, map to DTOs (never expose hash)

Private:
- **`generateTokenPair(user, tenant)`** — HS256 JWT with `eid`, `kind`, `iss:"pulse"`, `attributes.tenant_id`, `attributes.email`

Dependencies: `golang.org/x/crypto/bcrypt` (already in go.mod), `github.com/golang-jwt/jwt/v4` (already in go.mod), `github.com/google/uuid` (already in go.mod)

### TASK-022: Middleware

| File | Function | Behavior |
|---|---|---|
| `internal/auth/auth_middleware.go` | `AuthMiddleware(jwtSecret string) rest.Middleware` | Parse `Authorization: Bearer <token>` → `identity.FromToken()` → reject refresh tokens → `identity.WithContext()` |
| `internal/auth/tenant_middleware.go` | `TenantMiddleware() rest.Middleware` | Extract `tenant_id` from claims attributes → inject into context. Export `TenantIDFromContext(ctx) (int64, bool)` |
| `internal/auth/hmac_middleware.go` | `HMACMiddleware(hmacSecret string) rest.Middleware` | Verify `X-Signature: sha256=<hex>` against request body HMAC. Restore body for downstream. |

### TASK-023: Handler + Wiring

**Step 1.** Implement `internal/auth/handler.go`:
- `NewHandler(svc *AuthService) *Handler`
- `Routes(authMw, tenantMw rest.Middleware) []rest.Route` returns 5 routes:
  - `POST /v1/auth/login` — public
  - `POST /v1/auth/refresh` — public
  - `POST /v1/auth/keys` — protected (auth + tenant middleware)
  - `DELETE /v1/auth/keys/:id` — protected
  - `GET /v1/auth/keys` — protected
- Uses `httpx.ParseJsonBody` for request parsing
- Uses `response.OkJson` / `response.Error` for responses
- Path param via go-zero `httpx.GetPathParamFromContext` or string parsing from route `:id`

**Step 2.** Wire repos in `internal/registry/repository_context.go`:
- Add fields: `tenantRepo`, `userRepo`, `apiKeyRepo`, `auditLogRepo`
- Initialize in `NewRepositoryContext(db)` with `auth.NewXRepository(db)`
- Replace nil returns with concrete repo instances

**Step 3.** Wire real middleware in `internal/registry/security_http.go`:
- `NewHttpSecurityContext(c config.APIConfig)` — pass config instead of no-arg
- Use `auth.AuthMiddleware(c.Auth.JwtSecret)` and `auth.TenantMiddleware()`
- Update `security_context.go` to pass config (already does)

**Step 4.** Update `internal/handler/http_handler.go`:
- Create `RestHandler` struct with `Register(svr *rest.Server)` method
- Construct `AuthService` from `ServiceContext` repos + JWT secret
- Register auth routes on the server

**Step 5.** Update `cmd/api/main.go`:
- Replace `handler.NewHealthServer(...)` + unused `registry.NewServiceContext(c)` with:
  - Create `rest.Server` with health + auth routes
  - Use `NewServiceContext` to construct DI container
  - Wire `RestHandler.Register(srv)` for domain routes

### TASK-024: Smoke Test

Create `scripts/smoke-auth.sh` — 6 steps:
1. `POST /v1/auth/login` with seeded credentials → extract tokens
2. `POST /v1/auth/refresh` → verify token rotation
3. `POST /v1/auth/keys` → create key, extract raw_key and ID
4. `GET /v1/auth/keys` → verify key appears in list
5. `DELETE /v1/auth/keys/:id` → revoke the key
6. `GET /v1/auth/keys` (no auth) → verify 401

Response envelope is `{"meta":{...},"data":{...}}`, so jq extracts from `.data.access_token`.

---

## Files Changed

| File | Action | Task |
|---|---|---|
| `internal/contract/repository.go` | Add `UserRepo` interface | 020 |
| `internal/registry/repository_context.go` | Add `GetUserRepo()`, wire 4 auth repos | 020, 023 |
| `internal/service/service_factory.go` | Add `GetUserRepo()` to context interface | 020 |
| `internal/auth/tenant_repo.go` | Implement `TenantRepository` | 020 |
| `internal/auth/user_repo.go` | **New** — Implement `UserRepository` | 020 |
| `internal/auth/api_key_repo.go` | Implement `APIKeyRepository` | 020 |
| `internal/auth/audit_repo.go` | Implement `AuditLogRepository` | 020 |
| `internal/auth/dto.go` | **New** — Request/response DTOs | 021 |
| `internal/auth/service.go` | Implement `AuthService` (login, refresh, API key CRUD) | 021 |
| `internal/auth/auth_middleware.go` | JWT Bearer middleware | 022 |
| `internal/auth/tenant_middleware.go` | Tenant extraction middleware | 022 |
| `internal/auth/hmac_middleware.go` | HMAC signature middleware | 022 |
| `internal/auth/handler.go` | 5-endpoint REST handler + route registration | 023 |
| `internal/handler/http_handler.go` | `RestHandler.Register()` wiring | 023 |
| `internal/registry/security_http.go` | Replace passthroughs with real middleware | 023 |
| `internal/registry/security_context.go` | Pass config to `NewHttpSecurityContext` | 023 |
| `cmd/api/main.go` | Wire ServiceContext + RestHandler | 023 |
| `scripts/smoke-auth.sh` | **New** — End-to-end smoke test | 024 |

---

## Key Reusable Code

| What | Where |
|---|---|
| JWT parsing | `helper/utils/identity/parser.go` → `FromToken(token, nil, secret)` |
| Claims model | `helper/utils/identity/claims.go` → `MapClaims`, `Claims` interface |
| Claims→context | `helper/utils/identity/context.go` → `WithContext()`, `FromContext()` |
| Error types | `helper/utils/errors/helper.go` → `NewUnauthorized()`, `NewBadRequest()`, etc. |
| HTTP response | `helper/utils/server/http/response/helper.go` → `OkJson()`, `OkMsg()`, `Error()` |
| Outbox repo pattern | `internal/outbox/repository.go` — reference for GORM repo structure |
| Seed data | `scripts/seed.sh` — demo tenant/user/api-key for testing |

---

## Verification

1. `go build ./...` — all packages compile
2. `go test ./internal/auth/...` — unit tests pass (if added)
3. `docker compose up -d` → `make seed` → `make run` → `bash scripts/smoke-auth.sh` — end-to-end passes
4. Verify 401 on protected endpoints without token
5. Verify refresh token cannot be used as access token
6. Verify API key creation returns raw key, list does not expose hash
