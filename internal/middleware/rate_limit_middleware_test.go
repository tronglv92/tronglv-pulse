package middleware

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alicebob/miniredis/v2"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	zredis "github.com/zeromicro/go-zero/core/stores/redis"

	"pulse/helper/utils/stores/redis"
)

// newTestRedis starts a miniredis server and returns a Redis wrapper connected to it.
func newTestRedis(t *testing.T) (*redis.Redis, *miniredis.Miniredis) {
	t.Helper()
	mr := miniredis.RunT(t)
	rds, err := redis.NewRedis(redis.RedisConfig{
		RedisConf: zredis.RedisConf{
			Host:     mr.Addr(),
			Type:     "node",
			NonBlock: true,
		},
	})
	require.NoError(t, err)
	return rds, mr
}

func TestRateLimitMiddleware_AllowsUnderLimit(t *testing.T) {
	rds, _ := newTestRedis(t)

	mw := RateLimitMiddleware(rds, 10)
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rec := httptest.NewRecorder()

	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimitMiddleware_RejectsOverLimit(t *testing.T) {
	rds, _ := newTestRedis(t)

	rps := 5
	mw := RateLimitMiddleware(rds, rps)
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Exhaust the bucket (capacity = 5, so 5 requests allowed, 6th rejected).
	for i := 0; i < rps; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
		req.Header.Set("X-Tenant-ID", "tenant-exhaust")
		rec := httptest.NewRecorder()
		handler(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code, "request %d should pass", i+1)
	}

	// Next request should be rate-limited.
	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-exhaust")
	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)
	assert.Equal(t, "1", rec.Header().Get("Retry-After"))
}

func TestRateLimitMiddleware_FallsBackToRemoteAddr(t *testing.T) {
	rds, _ := newTestRedis(t)

	mw := RateLimitMiddleware(rds, 10)
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// No X-Tenant-ID header — should use RemoteAddr.
	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	rec := httptest.NewRecorder()

	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimitMiddleware_GracefulDegradation(t *testing.T) {
	rds, mr := newTestRedis(t)

	// Stop Redis to simulate unavailability.
	mr.Close()

	mw := RateLimitMiddleware(rds, 10)
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-1")
	rec := httptest.NewRecorder()

	// Should allow request through when Redis is down.
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestRateLimitMiddleware_IsolatesPerTenant(t *testing.T) {
	rds, _ := newTestRedis(t)

	rps := 2
	mw := RateLimitMiddleware(rds, rps)
	handler := mw(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
	})

	// Exhaust tenant-A's bucket.
	for i := 0; i < rps; i++ {
		req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
		req.Header.Set("X-Tenant-ID", "tenant-A")
		rec := httptest.NewRecorder()
		handler(rec, req)
		assert.Equal(t, http.StatusOK, rec.Code)
	}

	// tenant-A should be blocked.
	req := httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-A")
	rec := httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusTooManyRequests, rec.Code)

	// tenant-B should still pass.
	req = httptest.NewRequest(http.MethodGet, "/api/logs", nil)
	req.Header.Set("X-Tenant-ID", "tenant-B")
	rec = httptest.NewRecorder()
	handler(rec, req)
	assert.Equal(t, http.StatusOK, rec.Code)
}

func TestEvalTokenBucket_ReturnsAllowedOnFirstCall(t *testing.T) {
	rds, _ := newTestRedis(t)

	allowed, err := evalTokenBucket(context.Background(), rds, "ratelimit:test", 100)
	require.NoError(t, err)
	assert.True(t, allowed)
}
