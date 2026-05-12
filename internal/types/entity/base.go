package entity

import (
	"time"

	"gorm.io/gorm"
)

// Base is the standard base struct for most entities (BIGSERIAL PK, soft-delete).
type Base struct {
	ID        int64          `gorm:"primaryKey;autoIncrement"`
	CreatedAt time.Time      `gorm:"column:created_at;not null;default:now()"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:now()"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index"`
}

// HypertableBase is used by TimescaleDB hypertable entities.
// No DeletedAt — TimescaleDB chunk constraints conflict with partial-index soft-deletes.
// Composite PK (id, created_at) is required by TimescaleDB.
type HypertableBase struct {
	ID        int64     `gorm:"primaryKey;autoIncrement;column:id"`
	CreatedAt time.Time `gorm:"primaryKey;not null;default:now();column:created_at"`
}
