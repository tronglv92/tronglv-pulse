package jsonx

import (
	"encoding/json"
	"fmt"
)

// MustMarshalBytes marshals a value to JSON []byte, ignoring errors.
// Use only in scenarios where the value is known to be serializable.
func MustMarshalBytes(val any) []byte {
	b, _ := json.Marshal(val)
	return b
}

// MustMarshalString marshals a value to a JSON string, ignoring errors.
// Use for logging or debug output only.
func MustMarshalString(val any) string {
	return string(MustMarshalBytes(val))
}

// MapToStruct converts a map or JSON-like object into a typed struct using JSON round-trip.
// Useful for dynamic decoding from map[string]any or generic types.
func MapToStruct[T any](val any) (*T, error) {
	data, err := json.Marshal(val)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal value: %w", err)
	}

	var result T
	if err = json.Unmarshal(data, &result); err != nil {
		return nil, fmt.Errorf("failed to unmarshal to target struct: %w", err)
	}
	return &result, nil
}

// AnyToStruct converts any value to a struct of type T by JSON marshalling and unmarshalling.
func AnyToStruct[T any](val any) (*T, error) {
	b, err := json.Marshal(val)
	if err != nil {
		return nil, err
	}

	t := new(T)
	if err = json.Unmarshal(b, &t); err != nil {
		return nil, err
	}
	return t, nil
}
