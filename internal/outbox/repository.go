package outbox

import (
	"context"
	"fmt"
	"time"

	"pulse/internal/contract"
	"pulse/internal/types/define/enum"
	"pulse/internal/types/entity"

	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

var (
	_ contract.OutboxRepo     = (*Repository)(nil)
	_ contract.OutboxAppender = (*Repository)(nil)
)

// Repository implements contract.OutboxRepo and contract.OutboxAppender using GORM.
type Repository struct {
	db *gorm.DB
}

// NewRepository creates a new outbox Repository.
func NewRepository(db *gorm.DB) *Repository {
	return &Repository{db: db}
}

// Append inserts an outbox event within the caller's transaction.
// The caller must pass an active *gorm.DB transaction so the outbox row and
// the business write are committed atomically.
func (r *Repository) Append(ctx context.Context, tx *gorm.DB, e *entity.OutboxEvent) error {
	if err := tx.WithContext(ctx).Create(e).Error; err != nil {
		return fmt.Errorf("failed to append outbox event: %w", err)
	}
	return nil
}

// FindPending returns up to limit pending outbox events, locked with FOR UPDATE SKIP LOCKED
// to allow concurrent drainer workers without double-processing rows.
func (r *Repository) FindPending(ctx context.Context, limit int) ([]*entity.OutboxEvent, error) {
	var events []*entity.OutboxEvent

	err := r.db.WithContext(ctx).Transaction(func(tx *gorm.DB) error {
		return tx.
			Where("status = ?", enum.OutboxStatusPending).
			Order("created_at ASC").
			Limit(limit).
			Clauses(clause.Locking{Strength: "UPDATE", Options: "SKIP LOCKED"}).
			Find(&events).Error
	})
	if err != nil {
		return nil, fmt.Errorf("failed to find pending outbox events: %w", err)
	}

	return events, nil
}

// MarkPublished sets the event status to Published and records the publish timestamp.
func (r *Repository) MarkPublished(ctx context.Context, id int64) error {
	now := time.Now()
	result := r.db.WithContext(ctx).
		Model(&entity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":       enum.OutboxStatusPublished,
			"published_at": now,
			"updated_at":   now,
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark outbox event %d as published: %w", id, result.Error)
	}
	return nil
}

// MarkFailed sets the event status to Failed and increments the retry count.
func (r *Repository) MarkFailed(ctx context.Context, id int64, retryCount int16) error {
	result := r.db.WithContext(ctx).
		Model(&entity.OutboxEvent{}).
		Where("id = ?", id).
		Updates(map[string]any{
			"status":      enum.OutboxStatusFailed,
			"retry_count": retryCount,
			"updated_at":  time.Now(),
		})
	if result.Error != nil {
		return fmt.Errorf("failed to mark outbox event %d as failed: %w", id, result.Error)
	}
	return nil
}

// Delete permanently removes a published outbox row (no soft-delete on outbox).
func (r *Repository) Delete(ctx context.Context, id int64) error {
	result := r.db.WithContext(ctx).
		Where("id = ?", id).
		Delete(&entity.OutboxEvent{})
	if result.Error != nil {
		return fmt.Errorf("failed to delete outbox event %d: %w", id, result.Error)
	}
	return nil
}
