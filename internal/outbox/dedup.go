package outbox

import (
	"context"
	"fmt"
	"time"

	"pulse/internal/contract"
)

const dedupTTL = time.Hour

// Deduplicator prevents double-publishing of outbox events using Redis.
// Each published event ID is recorded as "evt:{id}" with a 1-hour TTL.
type Deduplicator struct {
	cache contract.Cache
}

// NewDeduplicator creates a Deduplicator backed by the given cache.
func NewDeduplicator(cache contract.Cache) *Deduplicator {
	return &Deduplicator{cache: cache}
}

// IsDuplicate returns true if the event has already been published.
func (d *Deduplicator) IsDuplicate(ctx context.Context, eventID int64) (bool, error) {
	return d.cache.Exists(ctx, dedupKey(eventID))
}

// Mark records the event ID in the cache so future IsDuplicate calls return true.
func (d *Deduplicator) Mark(ctx context.Context, eventID int64) error {
	return d.cache.Set(ctx, dedupKey(eventID), []byte("1"), dedupTTL)
}

func dedupKey(eventID int64) string {
	return fmt.Sprintf("evt:%d", eventID)
}
