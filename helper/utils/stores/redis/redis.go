package redis

import (
	"context"
	"errors"
	"fmt"
	"github.com/redis/go-redis/extra/redisotel/v9"
	red "github.com/redis/go-redis/v9"
	"github.com/zeromicro/go-zero/core/breaker"
	"github.com/zeromicro/go-zero/core/errorx"
	"github.com/zeromicro/go-zero/core/logx"
	zredis "github.com/zeromicro/go-zero/core/stores/redis"
	"github.com/zeromicro/go-zero/core/syncx"
	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/trace"
	"time"
)

const (
	NodeType = "node"
	// Nil is an alias of redis.Nil.
	Nil = red.Nil

	blockingQueryTimeout = 5 * time.Second
	readWriteTimeout     = 2 * time.Second
	defaultSlowThreshold = time.Millisecond * 100
	defaultPingTimeout   = time.Second
)

var (
	ErrNilNode    = errors.New("nil redis node")
	slowThreshold = syncx.ForAtomicDuration(defaultSlowThreshold)
)

type (
	RedisConfig struct {
		zredis.RedisConf
		Db int `json:"db,default=0"`
	}

	// Option defines the method to customize a Redis.
	Option func(r *Redis)

	// Redis defines a redis node/cluster. It is thread-safe.
	Redis struct {
		Addr           string
		Pass           string
		Db             int
		tls            bool
		brk            breaker.Breaker
		hooks          []red.Hook
		TraceProvider  trace.TracerProvider
		TraceOptions   []redisotel.TracingOption
		DisableTracing bool
	}
)

func WithDb(index int) Option {
	return func(r *Redis) {
		r.Db = index
	}
}

// WithPass customizes the given Redis with given password.
func WithPass(pass string) Option {
	return func(r *Redis) {
		r.Pass = pass
	}
}

// WithTLS customizes the given Redis with TLS enabled.
func WithTLS() Option {
	return func(r *Redis) {
		r.tls = true
	}
}

// withHook customizes the given Redis with given durationHook, only for private use now,
// maybe expose later.
func withHook(hook red.Hook) Option {
	return func(r *Redis) {
		r.hooks = append(r.hooks, hook)
	}
}

func WithDisableTracing() Option {
	return func(r *Redis) {
		r.DisableTracing = true
	}
}

func WithTraceProvider(provider trace.TracerProvider, opts ...redisotel.TracingOption) Option {
	return func(r *Redis) {
		r.TraceProvider = provider
		r.TraceOptions = opts
	}
}

func MustNewRedis(conf RedisConfig, opts ...Option) *Redis {
	rds, err := NewRedis(conf, opts...)
	logx.Must(err)
	return rds
}

func NewRedis(conf RedisConfig, opts ...Option) (*Redis, error) {
	if err := conf.Validate(); err != nil {
		return nil, err
	}

	if len(conf.Pass) > 0 {
		opts = append([]Option{WithPass(conf.Pass)}, opts...)
	}
	if conf.Tls {
		opts = append([]Option{WithTLS()}, opts...)
	}
	if conf.Db > 0 {
		opts = append([]Option{WithDb(conf.Db)}, opts...)
	}

	rds := newRedis(conf.Host, opts...)
	if !conf.NonBlock {
		if err := rds.checkConnection(conf.PingTimeout); err != nil {
			return nil, errorx.Wrap(err, fmt.Sprintf("redis connect error, addr: %s", conf.Host))
		}
	}
	return rds, nil
}

func newRedis(addr string, opts ...Option) *Redis {
	r := &Redis{
		Addr:          addr,
		brk:           breaker.NewBreaker(),
		TraceProvider: otel.GetTracerProvider(),
	}
	for _, opt := range opts {
		opt(r)
	}
	return r
}

func (s *Redis) checkConnection(pingTimeout time.Duration) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}

	timeout := defaultPingTimeout
	if pingTimeout > 0 {
		timeout = pingTimeout
	}

	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	return conn.Ping(ctx).Err()
}

func acceptable(err error) bool {
	return err == nil || errors.Is(err, red.Nil) || errors.Is(err, context.Canceled)
}

func (s *Redis) Del(keys ...string) (int, error) {
	return s.DelCtx(context.Background(), keys...)
}

func (s *Redis) DelCtx(ctx context.Context, keys ...string) (int, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}

	v, err := conn.Del(ctx, keys...).Result()
	if err != nil {
		return 0, err
	}

	return int(v), nil
}

func (s *Redis) Get(key string) (string, error) {
	return s.GetCtx(context.Background(), key)
}

func (s *Redis) GetCtx(ctx context.Context, key string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	if val, err := conn.Get(ctx, key).Result(); errors.Is(err, red.Nil) {
		return "", nil
	} else if err != nil {
		return "", err
	} else {
		return val, nil
	}
}

func (s *Redis) Set(key, value string) error {
	return s.SetCtx(context.Background(), key, value)
}

func (s *Redis) SetCtx(ctx context.Context, key, value string) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Set(ctx, key, value, 0).Err()
}

func (s *Redis) SetexCtx(ctx context.Context, key, value string, seconds int) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Set(ctx, key, value, time.Duration(seconds)*time.Second).Err()
}

func (s *Redis) SetnxExCtx(ctx context.Context, key, value string, seconds int) (bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return false, err
	}

	return conn.SetNX(ctx, key, value, time.Duration(seconds)*time.Second).Result()
}

func (s *Redis) LPush(ctx context.Context, key string, values ...any) error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.LPush(ctx, key, values...).Err()
}

func (s *Redis) Blpop(key string) (string, error) {
	return s.BlpopCtx(context.Background(), key)
}

func (s *Redis) BlpopCtx(ctx context.Context, key string) (string, error) {
	return s.BlpopWithTimeoutCtx(ctx, blockingQueryTimeout, key)
}

func (s *Redis) BlpopEx(key string) (string, bool, error) {
	return s.BlpopExCtx(context.Background(), key)
}

func (s *Redis) BlpopExCtx(ctx context.Context, key string) (string, bool, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", false, err
	}

	vals, err := conn.BLPop(ctx, blockingQueryTimeout, key).Result()
	if err != nil {
		return "", false, err
	}

	if len(vals) < 2 {
		return "", false, fmt.Errorf("no value on key: %s", key)
	}

	return vals[1], true, nil
}

func (s *Redis) BlpopWithTimeout(timeout time.Duration, key string) (string, error) {
	return s.BlpopWithTimeoutCtx(context.Background(), timeout, key)
}

func (s *Redis) BlpopWithTimeoutCtx(ctx context.Context, timeout time.Duration, key string) (string, error) {
	conn, err := getRedis(s)
	if err != nil {
		return "", err
	}

	vals, err := conn.BLPop(ctx, timeout, key).Result()
	if err != nil {
		return "", err
	}

	if len(vals) < 2 {
		return "", fmt.Errorf("no value on key: %s", key)
	}

	return vals[1], nil
}

// Scan is the implementation of redis scan command.
func (s *Redis) Scan(cursor uint64, match string, count int64) ([]string, uint64, error) {
	return s.ScanCtx(context.Background(), cursor, match, count)
}

// ScanCtx is the implementation of redis scan command.
func (s *Redis) ScanCtx(ctx context.Context, cursor uint64, match string, count int64) ([]string, uint64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, 0, err
	}
	return conn.Scan(ctx, cursor, match, count).Result()
}

func (s *Redis) ExistsCtx(ctx context.Context, keys ...string) (int64, error) {
	conn, err := getRedis(s)
	if err != nil {
		return 0, err
	}
	return conn.Exists(ctx, keys...).Result()
}

func (s *Redis) EvalCtx(ctx context.Context, script string, keys []string, args ...any) (any, error) {
	conn, err := getRedis(s)
	if err != nil {
		return nil, err
	}

	val, err := conn.Eval(ctx, script, keys, args...).Result()
	if err != nil && errors.Is(err, red.Nil) {
		return nil, nil
	}
	return val, err
}

func (s *Redis) Close() error {
	conn, err := getRedis(s)
	if err != nil {
		return err
	}
	return conn.Close()
}
