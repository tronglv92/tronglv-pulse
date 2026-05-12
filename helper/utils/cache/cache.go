package cache

import (
	"context"
	"pulse/helper/utils/errors"
	"pulse/helper/utils/stores/redis"
	"fmt"
	"github.com/zeromicro/go-zero/core/stores/cache"
	"github.com/zeromicro/go-zero/core/syncx"
	"log"
)

const (
	RedisDriver string = "redis"
)

type Config struct {
	Stack string            `json:"stack,default=redis"`
	Redis redis.RedisConfig `json:"redis,optional"`
}

type Cache interface {
	cache.Cache
	Exists(ctx context.Context, keys ...string) (int64, error)
	DelByPatternCtx(ctx context.Context, pattern string) error
}

func New(c Config, opts ...Option) Cache {
	switch c.Stack {
	case RedisDriver:
		return NewNode(
			redis.MustNewRedis(c.Redis),
			syncx.NewSingleFlight(),
			cache.NewStat(RedisDriver),
			errors.InternalServer(fmt.Errorf(RedisDriver)),
			opts...,
		)
	}
	log.Fatal("no cache driver support")
	return nil
}
