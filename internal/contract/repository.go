package contract

import (
	"context"
	"time"

	"github.com/pgvector/pgvector-go"
	"github.com/shopspring/decimal"
	"pulse/internal/types/define/enum"
	"pulse/internal/types/entity"
)

// LogFilter scopes log search and count queries.
type LogFilter struct {
	TenantID    int64
	ServiceName string
	Level       string
	Fingerprint string
	Query       string
	From        time.Time
	To          time.Time
	Page        int32
	Limit       int32
}

// LogRepo persists and queries log entries on the TimescaleDB hypertable.
type LogRepo interface {
	Insert(ctx context.Context, entries []*entity.LogEntry) error
	FindByID(ctx context.Context, id, tenantID int64) (*entity.LogEntry, error)
	Search(ctx context.Context, f LogFilter) ([]*entity.LogEntry, int64, error)
	// CountByWindow returns the log count for a service+fingerprint pair in [from, to).
	// Used by the anomaly detector to build rolling z-score windows.
	CountByWindow(ctx context.Context, tenantID int64, service, fingerprint string, from, to time.Time) (int64, error)
}

// IncidentFilter scopes incident list queries.
type IncidentFilter struct {
	TenantID    int64
	Status      enum.IncidentStatus
	ServiceName string
	Page        int32
	Limit       int32
}

// IncidentRepo handles CRUD for the incidents table.
type IncidentRepo interface {
	Create(ctx context.Context, e *entity.Incident) error
	FindByID(ctx context.Context, id, tenantID int64) (*entity.Incident, error)
	FindByFingerprint(ctx context.Context, tenantID int64, fingerprint string) (*entity.Incident, error)
	Update(ctx context.Context, e *entity.Incident) error
	List(ctx context.Context, f IncidentFilter) ([]*entity.Incident, int64, error)
}

// EmbeddingRepo stores and retrieves pgvector error-fingerprint embeddings.
type EmbeddingRepo interface {
	Upsert(ctx context.Context, e *entity.ErrorEmbedding) error
	FindByFingerprint(ctx context.Context, tenantID int64, fingerprint string) (*entity.ErrorEmbedding, error)
	// FindSimilar returns up to topK embeddings ordered by cosine distance to vec.
	FindSimilar(ctx context.Context, tenantID int64, vec pgvector.Vector, topK int) ([]*entity.ErrorEmbedding, error)
}

// LLMCallRepo records every LLM API invocation for cost tracking.
type LLMCallRepo interface {
	Save(ctx context.Context, e *entity.LLMCall) error
	// SumCostSince returns total USD spend across all tenants since the given time.
	// Used by the cron budget-guard to enforce the daily cap.
	SumCostSince(ctx context.Context, since time.Time) (decimal.Decimal, error)
}

// TenantRepo manages tenant records.
type TenantRepo interface {
	Save(ctx context.Context, e *entity.Tenant) error
	FindByID(ctx context.Context, id int64) (*entity.Tenant, error)
	FindBySlug(ctx context.Context, slug string) (*entity.Tenant, error)
	Update(ctx context.Context, e *entity.Tenant) error
}

// APIKeyRepo supports the HMAC API-key authentication middleware.
type APIKeyRepo interface {
	Save(ctx context.Context, e *entity.APIKey) error
	// FindByHash looks up an active key by its full HMAC hash (field key_hash).
	FindByHash(ctx context.Context, keyHash string) (*entity.APIKey, error)
	ListByTenant(ctx context.Context, tenantID int64) ([]*entity.APIKey, error)
	Delete(ctx context.Context, id, tenantID int64) error
	// TouchLastUsed records the access timestamp without a full entity update.
	TouchLastUsed(ctx context.Context, id int64) error
}

// OutboxRepo is the drainer-side view of the outbox table.
// To write within a DB transaction use OutboxAppender (outbox.go).
type OutboxRepo interface {
	FindPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, retryCount int16) error
	// Delete permanently removes a successfully published row (no soft-delete).
	Delete(ctx context.Context, id int64) error
}

// SagaRepo provides basic CRUD for saga orchestration state.
type SagaRepo interface {
	Create(ctx context.Context, e *entity.SagaInstance) error
	FindByID(ctx context.Context, id int64) (*entity.SagaInstance, error)
	Update(ctx context.Context, e *entity.SagaInstance) error
	FindPending(ctx context.Context) ([]*entity.SagaInstance, error)
}

// UserRepo manages user records for authentication.
type UserRepo interface {
	FindByEmail(ctx context.Context, email string) (*entity.User, error)
	FindByID(ctx context.Context, id int64) (*entity.User, error)
}

// AuditLogRepo is append-only — never updates or deletes rows.
type AuditLogRepo interface {
	Append(ctx context.Context, e *entity.AuditLog) error
	FindByResource(ctx context.Context, tenantID int64, resource string, resourceID int64) ([]*entity.AuditLog, error)
}
