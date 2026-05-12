package cache

import (
	"context"
	"pulse/helper/utils/toolkit/contextx"
	"pulse/helper/utils/toolkit/stringx"
	"fmt"
	"strings"
	"time"
)

const keySeparator = ","

func Key(params ...any) string {
	s := "inapp"
	for _, v := range params {
		s = fmt.Sprintf("%s:%v", s, v)
	}
	return s
}

func KeyCtx(ctx context.Context, params ...any) string {
	serviceName := contextx.MustValue[string](ctx, contextx.KeyServiceName)
	if len(serviceName) > 0 {
		serviceName = strings.Replace(stringx.Slugify(serviceName), "-", "_", -1)
		params = append([]any{serviceName}, params...)
	}
	return Key(params...)
}

func RememberCtx[T any](ctx context.Context, client Cache, key string, expire time.Duration, query func() (any, error)) (T, error) {
	var result T
	if err := client.GetCtx(ctx, key, &result); err == nil {
		return result, nil
	}

	raw, err := query()
	if err != nil {
		return result, err
	}

	val, ok := raw.(T)
	if !ok {
		return result, fmt.Errorf("RememberCtx: query returned wrong type: expected %T, got %T", result, raw)
	}

	if err = client.SetWithExpireCtx(ctx, key, val, expire); err != nil {
		return val, err
	}
	return val, nil
}

func formatKeys(keys []string) string {
	return strings.Join(keys, keySeparator)
}
