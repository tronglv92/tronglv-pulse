package redis

import (
	"crypto/tls"
	"github.com/redis/go-redis/extra/redisotel/v9"
	red "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/logx"
	"github.com/zeromicro/go-zero/core/syncx"
	"go.opentelemetry.io/otel"
	"io"
	"runtime"
)

const (
	defaultDatabase = 0
	maxRetries      = 3
	idleConns       = 8
)

var (
	clientManager = syncx.NewResourceManager()
	nodePoolSize  = 10 * runtime.GOMAXPROCS(0)
)

func getRedis(r *Redis) (*red.Client, error) {
	val, err := clientManager.GetResource(r.Addr, func() (io.Closer, error) {
		var tlsConfig *tls.Config
		if r.tls {
			tlsConfig = &tls.Config{
				InsecureSkipVerify: true,
			}
		}
		store := red.NewClient(&red.Options{
			Addr:         r.Addr,
			Password:     r.Pass,
			DB:           r.Db,
			MaxRetries:   maxRetries,
			MinIdleConns: idleConns,
			TLSConfig:    tlsConfig,
		})

		hooks := append([]red.Hook{defaultDurationHook, breakerHook{
			brk: r.brk,
		}}, r.hooks...)
		for _, hook := range hooks {
			store.AddHook(hook)
		}

		connCollector.registerClient(&statGetter{
			clientType: NodeType,
			key:        r.Addr,
			poolSize:   nodePoolSize,
			poolStats: func() *red.PoolStats {
				return store.PoolStats()
			},
		})

		opts := append([]redisotel.TracingOption{redisotel.WithTracerProvider(otel.GetTracerProvider())})
		if err := redisotel.InstrumentTracing(store, opts...); err != nil {
			logx.Error(err)
		}
		return store, nil
	})
	if err != nil {
		return nil, err
	}
	return val.(*red.Client), nil
}
