package filter

import (
	"pulse/helper/utils/toolkit/numberx"
	"google.golang.org/protobuf/types/known/timestamppb"
	"reflect"
	"strings"
	"time"
)

func BuildMap[T any](req T, base map[string]any, opts ...SMOption) map[string]any {
	return NewStructToMap(opts...).ConvertWithAppend(req, base)
}

func BuildJSONMap[T any](req T, base map[string]any) map[string]any {
	return BuildMap(req, base, WithTag("json"))
}

func BuildFilterMap[T any](req T, base map[string]any) map[string]any {
	return BuildMap(req, base)
}

func Extract[T any](m map[string]any, key string) (T, bool) {
	var result T
	if val, ok := m[key]; ok {
		if typedVal, ok := val.(T); ok {
			return typedVal, true
		}
	}
	return result, false
}

func ExtractNonEmptyString(m map[string]any, key string) (string, bool) {
	if val, ok := Extract[string](m, key); ok && len(val) > 0 {
		return val, true
	}
	return "", false
}

func ExtractPositiveInt(m map[string]any, key string) (int64, bool) {
	if val, ok := m[key]; ok {
		if v := ToUnsignedInt(val); v > 0 {
			return v, true
		}
	}
	return 0, false
}

func ExtractUnsignedInt(m map[string]any, key string) (int64, bool) {
	if val, ok := m[key]; ok {
		if v := ToUnsignedInt(val); v >= 0 {
			return v, true
		}
	}
	return 0, false
}

func ContainsNonZeroKey(m map[string]any, keys ...string) bool {
	for _, k := range keys {
		if val, ok := m[k]; ok && !reflect.ValueOf(val).IsZero() {
			return true
		}
	}
	return false
}

func ExtractNonZeroValue(m map[string]any, key string) (any, bool) {
	if val, ok := m[key]; ok && !reflect.ValueOf(val).IsZero() {
		return val, true
	}
	return nil, false
}

func ExtractIdList(input string) []int {
	return numberx.SplitToInts(input, ",")
}

func IsNumeric(v any) bool {
	switch v.(type) {
	case int, int8, int16, int32, int64,
		uint, uint8, uint16, uint32, uint64,
		float32, float64:
		return true
	default:
		return false
	}
}

func ToUnsignedInt(v any) int64 {
	switch n := v.(type) {
	case int:
		return int64(n)
	case int8:
		return int64(n)
	case int16:
		return int64(n)
	case int32:
		return int64(n)
	case int64:
		return n
	case uint:
		return int64(n)
	case uint8:
		return int64(n)
	case uint16:
		return int64(n)
	case uint32:
		return int64(n)
	case uint64:
		return int64(n)
	case float32:
		return int64(n)
	case float64:
		return int64(n)
	default:
		val := reflect.ValueOf(v)
		switch val.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			return val.Int()
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			return int64(val.Uint())
		case reflect.Float32, reflect.Float64:
			return int64(val.Float())
		default:
			return -1
		}
	}
}

func ExtractBool(m map[string]any, key string) (bool, bool) {
	val, exists := m[key]
	if !exists || val == nil {
		return false, false
	}

	switch v := val.(type) {
	case bool:
		return v, true
	case string:
		switch strings.ToLower(v) {
		case "true":
			return true, true
		case "false":
			return false, true
		}
	case float64, int:
		if v == 1 {
			return true, true
		}
		if v == 0 {
			return false, true
		}
		return false, false
	}
	return false, false
}

func ExtractProtoTimestamp(m map[string]any, key string) (time.Time, bool) {
	if val, ok := m[key]; ok && val != nil {
		switch v := val.(type) {
		case *timestamppb.Timestamp:
			if v != nil {
				return v.AsTime(), true
			}
		case time.Time:
			return v, true
		case string:
			if t, err := time.Parse(time.RFC3339, v); err == nil {
				return t, true
			}
			if t, err := time.Parse("2006-01-02 15:04:05", v); err == nil {
				return t, true
			}
		}
	}
	return time.Time{}, false
}
