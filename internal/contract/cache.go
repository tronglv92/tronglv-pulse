package contract

import (
	"context"
	"time"
)

// Cache is a generic key-value store backed by Redis.
// Values are raw bytes; callers are responsible for JSON encoding/decoding.
// Get returns (nil, nil) on a cache miss.
type Cache interface {
	Get(ctx context.Context, key string) ([]byte, error)
	// Set stores value with ttl. ttl=0 means no expiry.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	Del(ctx context.Context, keys ...string) error
	// Exists returns true when all supplied keys are present.
	// Used by the LLM budget guard to check llm:disabled.
	Exists(ctx context.Context, keys ...string) (bool, error)
}
