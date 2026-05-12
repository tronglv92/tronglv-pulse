package entity

import (
	"encoding/json"
	"time"

	"pulse/internal/types/define/enum"
)

// OutboxEvent implements the Transactional Outbox pattern (ADR-008).
// Never publish to Kafka directly from a service that also writes to the DB —
// write here in the same transaction, then let the drainer publish.
type OutboxEvent struct {
	ID          int64            `gorm:"primaryKey;autoIncrement;column:id"`
	AggregateID string           `gorm:"column:aggregate_id;not null;size:255;index"`
	EventType   string           `gorm:"column:event_type;not null;size:100"`
	Topic       string           `gorm:"column:topic;not null;size:255"`
	Payload     json.RawMessage  `gorm:"column:payload;type:jsonb;not null"`
	Status      enum.OutboxStatus `gorm:"column:status;not null;default:0;index"`
	RetryCount  int16            `gorm:"column:retry_count;not null;default:0"`
	PublishedAt *time.Time       `gorm:"column:published_at"`
	CreatedAt   time.Time        `gorm:"column:created_at;not null;default:now();index"`
	UpdatedAt   time.Time        `gorm:"column:updated_at;not null;default:now()"`
}

func (OutboxEvent) TableName() string { return "outbox_events" }
