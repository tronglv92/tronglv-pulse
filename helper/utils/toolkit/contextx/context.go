package contextx

import (
	"context"
)

type contextKey string

const (
	authorizationKey contextKey = "authorization"

	KeyElapsedTimer = "contextx:elapsed-timer"
	KeyServiceEnv   = "contextx:service-env"
	KeyServiceName  = "contextx:service-name"
)

func WithValue(ctx context.Context, key string, value any) context.Context {
	return context.WithValue(ctx, key, value)
}

func Value[T any](ctx context.Context, key string) (T, bool) {
	val, ok := ctx.Value(key).(T)
	return val, ok
}

func MustValue[T any](ctx context.Context, key string) T {
	val, ok := Value[T](ctx, key)
	if !ok {
		var zero T
		return zero
	}
	return val
}
