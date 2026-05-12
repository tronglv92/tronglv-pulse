package localize

import (
	"context"
)

func WithLangContext(parent context.Context, lang string) context.Context {
	return context.WithValue(parent, LanguageKey, lang)
}

func WithContext(parent context.Context, localizer Localizer) context.Context {
	return context.WithValue(parent, LocalizeKey, localizer)
}

func FromContext(ctx context.Context) Localizer {
	if localizer, ok := ctx.Value(LocalizeKey).(Localizer); ok {
		return localizer
	}
	return NewWithDefault()
}

func GetString(ctx context.Context, key string) string {
	return FromContext(ctx).GetString(key)
}

func GetFormat(ctx context.Context, key string, v ...any) string {
	return FromContext(ctx).GetFormat(key, v...)
}
