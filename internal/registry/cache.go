package registry

import (
	"context"
	"fmt"
	"time"

	helperredis "pulse/helper/utils/stores/redis"
	"pulse/internal/config"
	"pulse/internal/contract"

	zredis "github.com/zeromicro/go-zero/core/stores/redis"
)

var _ contract.Cache = (*redisCache)(nil)

// redisCache adapts helper/utils/stores/redis (string-based API) to contract.Cache ([]byte API).
type redisCache struct {
	rds *helperredis.Redis
}

func newRedisCache(c config.CacheConfig) *redisCache {
	rds := helperredis.MustNewRedis(helperredis.RedisConfig{
		RedisConf: zredis.RedisConf{
			Host:     c.Host,
			Type:     helperredis.NodeType,
			Pass:     c.Pass,
			NonBlock: true,
		},
	})
	return &redisCache{rds: rds}
}

func (r *redisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := r.rds.GetCtx(ctx, key)
	if err != nil {
		return nil, fmt.Errorf("failed to get key %s: %w", key, err)
	}
	if val == "" {
		return nil, nil // cache miss
	}
	return []byte(val), nil
}

func (r *redisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	if ttl == 0 {
		return r.rds.SetCtx(ctx, key, string(value))
	}
	seconds := int(ttl.Seconds())
	if seconds < 1 {
		seconds = 1
	}
	return r.rds.SetexCtx(ctx, key, string(value), seconds)
}

func (r *redisCache) Del(ctx context.Context, keys ...string) error {
	_, err := r.rds.DelCtx(ctx, keys...)
	return err
}

func (r *redisCache) Exists(ctx context.Context, keys ...string) (bool, error) {
	n, err := r.rds.ExistsCtx(ctx, keys...)
	if err != nil {
		return false, err
	}
	return n == int64(len(keys)), nil
}
