package contract

import (
	"context"

	"gorm.io/gorm"
	"pulse/internal/types/entity"
)

// OutboxAppender is the write side of the Transactional Outbox pattern (ADR-008).
// MUST be called with an active *gorm.DB transaction so the outbox row and the
// business write are committed atomically. Use TxManager.RunInTx to obtain tx.
type OutboxAppender interface {
	Append(ctx context.Context, tx *gorm.DB, e *entity.OutboxEvent) error
}

// OutboxRepository is the drainer's view of the outbox table.
// FindPending implementations must use FOR UPDATE SKIP LOCKED to allow
// concurrent drainer workers without double-processing rows.
type OutboxRepository interface {
	FindPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error)
	MarkPublished(ctx context.Context, id int64) error
	MarkFailed(ctx context.Context, id int64, retryCount int16) error
	// Delete permanently removes a published row (no soft-delete on outbox).
	Delete(ctx context.Context, id int64) error
}
