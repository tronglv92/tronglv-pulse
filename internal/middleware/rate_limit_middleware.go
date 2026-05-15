package middleware

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/rest"

	"pulse/helper/utils/errors"
	"pulse/helper/utils/server/http/response"
	"pulse/helper/utils/stores/redis"
)

// tokenBucketScript is a Lua script implementing an atomic token-bucket algorithm.
// KEYS[1] = rate limit key
// ARGV[1] = bucket capacity (max tokens / requests per second)
// ARGV[2] = current unix timestamp in seconds (float)
// ARGV[3] = key TTL in seconds
//
// Returns 1 if allowed, 0 if rejected.
const tokenBucketScript = `
local key = KEYS[1]
local capacity = tonumber(ARGV[1])
local now = tonumber(ARGV[2])
local ttl = tonumber(ARGV[3])

local bucket = redis.call("HMGET", key, "tokens", "last_refill")
local tokens = tonumber(bucket[1])
local last_refill = tonumber(bucket[2])

if tokens == nil then
    tokens = capacity
    last_refill = now
end

local elapsed = now - last_refill
local refill = elapsed * capacity
tokens = math.min(capacity, tokens + refill)

if tokens < 1 then
    return 0
end

tokens = tokens - 1
redis.call("HMSET", key, "tokens", tokens, "last_refill", now)
redis.call("EXPIRE", key, ttl)
return 1
`

const rateLimitKeyTTL = 2 // seconds — auto-cleanup for idle tenants

// RateLimitMiddleware returns a rest.Middleware that enforces a per-tenant
// token-bucket rate limit backed by Redis. If Redis is unavailable the request
// is allowed through (graceful degradation).
//
// Tenant identification order:
//  1. X-Tenant-ID request header (set by upstream auth/HMAC middleware)
//  2. Client IP address (fallback for unauthenticated requests)
func RateLimitMiddleware(rds *redis.Redis, rps int) rest.Middleware {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			ctx := r.Context()

			tenant := r.Header.Get("X-Tenant-ID")
			if tenant == "" {
				tenant = r.RemoteAddr
			}

			key := fmt.Sprintf("ratelimit:%s", tenant)

			allowed, err := evalTokenBucket(ctx, rds, key, rps)
			if err != nil {
				// Graceful degradation: allow request if Redis is unavailable.
				logx.WithContext(ctx).Errorw("rate limiter redis error, allowing request",
					logx.Field("error", err.Error()),
					logx.Field("tenant", tenant),
				)
				next(w, r)
				return
			}

			if !allowed {
				w.Header().Set("Retry-After", "1")
				response.Error(ctx, w, errors.New(
					http.StatusTooManyRequests,
					"RATE_LIMIT_EXCEEDED",
					"Too many requests. Please try again later.",
				))
				return
			}

			next(w, r)
		}
	}
}

// evalTokenBucket executes the token-bucket Lua script against Redis.
// Returns true if the request is allowed, false if rate-limited.
func evalTokenBucket(ctx context.Context, rds *redis.Redis, key string, rps int) (bool, error) {
	now := float64(time.Now().UnixNano()) / 1e9

	result, err := rds.EvalCtx(ctx, tokenBucketScript, []string{key}, rps, now, rateLimitKeyTTL)
	if err != nil {
		return false, fmt.Errorf("failed to eval rate limit script: %w", err)
	}

	allowed, ok := result.(int64)
	if !ok {
		return false, fmt.Errorf("unexpected rate limit script result type: %T", result)
	}

	return allowed == 1, nil
}
