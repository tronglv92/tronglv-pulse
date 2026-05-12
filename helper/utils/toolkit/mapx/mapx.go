package mapx

import (
	"maps"
)

// GetStringFromMap retrieves the string value for the given key from a map[string]string.
// Returns an empty string if the key does not exist.
func GetStringFromMap(m map[string]string, key string) string {
	if v, ok := m[key]; ok {
		return v
	}
	return ""
}

// GetValueFromAnyMap safely retrieves a value of type T from map[string]any by key.
// Returns the value and true if found and type matches, or zero value and false otherwise.
func GetValueFromAnyMap[T any](m map[string]any, key string) (T, bool) {
	if v, ok := m[key]; ok {
		if t, ok := v.(T); ok {
			return t, true
		}
	}
	var zero T
	return zero, false
}

// Merge copies all key-value pairs from map mx into map m.
// If mx is nil, m is returned unchanged.
// If a key from mx already exists in m, its value will be overwritten.
// The resulting map m is returned.
func Merge[K comparable, V any](m, mx map[K]V) map[K]V {
	if mx != nil {
		maps.Copy(m, mx)
	}
	return m
}
