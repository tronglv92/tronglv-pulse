package contextx

import (
	"context"
	"google.golang.org/grpc/metadata"
)

func WithAuthorizationHeader(ctx context.Context, token string) context.Context {
	return WithOutgoingContext(ctx, string(authorizationKey), token)
}

func WithAuthorizationFromContext(ctx context.Context) context.Context {
	return WithOutgoingContext(ctx, string(authorizationKey), AuthorizationFromContext(ctx))
}

func WithOutgoingContext(ctx context.Context, kv ...string) context.Context {
	return metadata.AppendToOutgoingContext(ctx, kv...)
}

func WithAuthorizationToken(ctx context.Context, token string) context.Context {
	return context.WithValue(ctx, authorizationKey, token)
}

func AuthorizationFromContext(ctx context.Context) string {
	token, ok := ctx.Value(authorizationKey).(string)
	if !ok {
		return ""
	}
	return token
}
