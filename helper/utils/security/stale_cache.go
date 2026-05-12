package security

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/zeromicro/go-zero/core/collection"
	"github.com/zeromicro/go-zero/core/logx"
)

const (
	// StaleRoleKey is the cache key for stale role-permission mappings
	StaleRoleKey = "security:roles:stale"

	// StaleMetadataKey stores metadata about when the stale cache was last updated
	StaleMetadataKey = "security:roles:stale:meta"

	// StaleCacheTTL is the maximum age for stale cache data (24 hours)
	StaleCacheTTL = 24 * time.Hour
)

// CacheMetadata tracks when cached data was stored and when it expires.
type CacheMetadata struct {
	CachedAt  time.Time `json:"cached_at"`
	ExpiresAt time.Time `json:"expires_at"`
}

// StaleAwareCache wraps go-zero's collection.Cache with stale cache fallback logic.
// It maintains two cache layers:
// 1. Fresh cache with normal TTL (6 hours)
// 2. Stale cache with extended TTL (24 hours) as fallback
//
// When fresh cache fails to load, it falls back to stale cache if the data
// is still within the acceptable staleness window (24 hours).
type StaleAwareCache struct {
	freshCache *collection.Cache // Primary cache with normal TTL
	staleCache *collection.Cache // Backup cache with extended TTL
	staleTTL   time.Duration     // Maximum age for stale data
}

// NewStaleAwareCache creates a new stale-aware cache wrapper.
//
// Parameters:
//   - freshCache: Primary cache for fresh data
//   - staleCache: Secondary cache for backup data with longer TTL
//
// The staleCache should be configured with a TTL equal to or greater than StaleCacheTTL.
func NewStaleAwareCache(freshCache, staleCache *collection.Cache) *StaleAwareCache {
	return &StaleAwareCache{
		freshCache: freshCache,
		staleCache: staleCache,
		staleTTL:   StaleCacheTTL,
	}
}

// TakeWithStale attempts to get data from fresh cache first, with stale cache fallback.
//
// Returns:
//   - data: The cached data (from fresh or stale cache)
//   - isStale: true if data came from stale cache, false if fresh
//   - error: Non-nil if both fresh and stale cache failed, or stale data too old
//
// Behavior:
// 1. Try fresh cache first
// 2. On success: Update stale cache as backup, return fresh data
// 3. On failure: Check stale cache
// 4. If stale data exists and age < 24 hours: Return stale data with isStale=true
// 5. If stale data too old or missing: Return error
func (s *StaleAwareCache) TakeWithStale(ctx context.Context, key string, fetch func(ctx context.Context) (any, error)) (any, bool, error) {
	// Try fresh cache first
	val, err := s.freshCache.Take(key, func() (any, error) {
		data, fetchErr := fetch(ctx)
		if fetchErr == nil {
			// Success - update stale cache as backup
			s.updateStaleCache(ctx, data)
		}
		return data, fetchErr
	})

	if err == nil {
		return val, false, nil // Fresh cache hit
	}

	// Fresh cache failed - try stale cache
	logx.WithContext(ctx).Errorf("fresh cache failed (%v), attempting stale cache fallback", err)

	staleData, staleErr := s.getStaleData(ctx)
	if staleErr != nil {
		return nil, false, fmt.Errorf("both fresh and stale cache failed: fresh=%v, stale=%v", err, staleErr)
	}

	// Check staleness metadata
	meta, metaErr := s.getStaleMetadata(ctx)
	if metaErr != nil {
		logx.WithContext(ctx).Errorf("stale cache metadata missing: %v, using stale data anyway", metaErr)
	} else {
		age := time.Since(meta.CachedAt)
		if age > s.staleTTL {
			return nil, false, fmt.Errorf("stale cache too old (age: %v, limit: %v)", age, s.staleTTL)
		}

		logx.WithContext(ctx).Infof("using stale cache (age: %v, expires in: %v)",
			age, time.Until(meta.ExpiresAt))
	}

	return staleData, true, nil // Stale cache hit
}

// updateStaleCache stores fresh data in the stale cache as a backup.
func (s *StaleAwareCache) updateStaleCache(ctx context.Context, data any) {
	meta := CacheMetadata{
		CachedAt:  time.Now(),
		ExpiresAt: time.Now().Add(s.staleTTL),
	}

	// Serialize data for storage
	dataBytes, err := json.Marshal(data)
	if err != nil {
		logx.WithContext(ctx).Errorf("Failed to serialize data for stale cache: %v", err)
		return
	}

	// Store data in stale cache
	_, err = s.staleCache.Take(StaleRoleKey, func() (any, error) {
		return string(dataBytes), nil
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("Failed to update stale cache data: %v", err)
	}

	// Store metadata
	metaBytes, err := json.Marshal(meta)
	if err != nil {
		logx.WithContext(ctx).Errorf("Failed to serialize metadata for stale cache: %v", err)
		return
	}

	_, err = s.staleCache.Take(StaleMetadataKey, func() (any, error) {
		return string(metaBytes), nil
	})
	if err != nil {
		logx.WithContext(ctx).Errorf("Failed to update stale cache metadata: %v", err)
	}
}

// getStaleData retrieves data from the stale cache.
func (s *StaleAwareCache) getStaleData(ctx context.Context) (map[string][]string, error) {
	val, err := s.staleCache.Take(StaleRoleKey, func() (any, error) {
		return nil, fmt.Errorf("stale cache miss")
	})
	if err != nil {
		return nil, err
	}

	dataStr, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("stale cache data has invalid type: %T", val)
	}

	var data map[string][]string
	if err := json.Unmarshal([]byte(dataStr), &data); err != nil {
		return nil, fmt.Errorf("failed to deserialize stale cache data: %w", err)
	}

	return data, nil
}

// getStaleMetadata retrieves metadata about the stale cache.
func (s *StaleAwareCache) getStaleMetadata(ctx context.Context) (*CacheMetadata, error) {
	val, err := s.staleCache.Take(StaleMetadataKey, func() (any, error) {
		return nil, fmt.Errorf("stale metadata miss")
	})
	if err != nil {
		return nil, err
	}

	metaStr, ok := val.(string)
	if !ok {
		return nil, fmt.Errorf("stale metadata has invalid type: %T", val)
	}

	var meta CacheMetadata
	if err := json.Unmarshal([]byte(metaStr), &meta); err != nil {
		return nil, fmt.Errorf("failed to deserialize stale metadata: %w", err)
	}

	return &meta, nil
}
