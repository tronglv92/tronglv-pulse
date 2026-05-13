package contract

import (
	"context"

	"gorm.io/gorm"
)

// TxManager executes a function within a single DB transaction.
// Return a non-nil error from fn to roll back; nil commits.
// The *gorm.DB passed to fn is the active transaction — pass it to
// OutboxAppender.Append and any other write that must be atomic.
type TxManager interface {
	RunInTx(ctx context.Context, fn func(tx *gorm.DB) error) error
}
