package entity

import (
	"encoding/json"
	"time"
)

// AuditLog is append-only — no UpdatedAt, no DeletedAt.
// DB-level immutability is enforced via Postgres RULEs (see migration 0007).
type AuditLog struct {
	ID         int64           `gorm:"primaryKey;autoIncrement;column:id"`
	TenantID   int64           `gorm:"column:tenant_id;not null;index"`
	ActorID    *int64          `gorm:"column:actor_id;index"`
	ActorType  string          `gorm:"column:actor_type;size:50"`
	Action     string          `gorm:"column:action;not null;size:100"`
	Resource   string          `gorm:"column:resource;not null;size:100"`
	ResourceID *int64          `gorm:"column:resource_id"`
	Metadata   json.RawMessage `gorm:"column:metadata;type:jsonb"`
	CreatedAt  time.Time       `gorm:"column:created_at;not null;default:now()"`
}

func (AuditLog) TableName() string { return "audit_logs" }
