package db

import (
	"context"
	"fmt"

	"pulse/helper/utils/cache"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type redisAdapter struct {
	config RedisConfig
	client *redis.Client
	cache  cache.Cache
}

func MapRedisAdapter(c BaseConfig, opts ...Opt) (Adapter, error) {
	pc, ok := c.(RedisConfig)
	if !ok {
		return nil, fmt.Errorf("invalid config for redis adapter")
	}
	return NewRedisAdapter(pc, opts...)
}

func NewRedisAdapter(config RedisConfig, opts ...Opt) (Adapter, error) {
	return &redisAdapter{
		config: config,
	}, nil
}

func (r *redisAdapter) Connect(ctx context.Context) error {
	return nil
}

func (r *redisAdapter) Close() error {
	if r.client != nil {
		return r.client.Close()
	}
	return nil
}

func (r *redisAdapter) GetDB() *gorm.DB {
	return nil
}

func (r *redisAdapter) GetCache() cache.Cache {
	return r.cache
}
