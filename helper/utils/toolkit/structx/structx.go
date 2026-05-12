package structx

import "google.golang.org/protobuf/types/known/structpb"

// PBToMap converts a map of protobuf Struct values (map[string]*structpb.Value)
// into a native Go map[string]any. It recursively handles nested Structs and lists.
func PBToMap(m map[string]*structpb.Value) map[string]any {
	result := make(map[string]any, len(m))
	for k, v := range m {
		result[k] = convertValue(v)
	}
	return result
}

// convertValue transforms a single *structpb.Value into its corresponding native Go type.
// Supported types include:
// - string
// - float64 (for numbers)
// - bool
// - nil (for null values)
// - map[string]any (for nested Structs)
// - []any (for lists)
func convertValue(v *structpb.Value) any {
	switch kind := v.Kind.(type) {
	case *structpb.Value_NullValue:
		return nil
	case *structpb.Value_StringValue:
		return kind.StringValue
	case *structpb.Value_NumberValue:
		return kind.NumberValue
	case *structpb.Value_BoolValue:
		return kind.BoolValue
	case *structpb.Value_StructValue:
		return PBToMap(kind.StructValue.Fields)
	case *structpb.Value_ListValue:
		values := make([]any, len(kind.ListValue.Values))
		for i, item := range kind.ListValue.Values {
			values[i] = convertValue(item)
		}
		return values
	default:
		return nil
	}
}

// PrepareStructList converts a slice of any type T into a *structpb.Value (a protobuf list value),
// where each item is transformed into a protobuf struct using the provided mapping function `m`.
//
// Parameters:
//   - items: A slice of any type T (can be struct or pointer to struct).
//   - m: A function that maps each item T into a map[string]any, which is then converted to *structpb.Struct.
//
// Returns:
//   - *structpb.Value representing a list of protobuf Struct values,
//     or nil if any conversion fails.
//
// Example:
//
//	PrepareStructList(ctx, users, func(u User) map[string]any {
//	    return map[string]any{
//	        "EmployeeId": u.EmployeeId,
//	        "Name":       u.Name,
//	    }
//	})
func PrepareStructList[T any](items []T, m func(T) map[string]any) *structpb.Value {
	results := make([]*structpb.Value, 0, len(items))
	for _, item := range items {
		s, err := structpb.NewStruct(m(item))
		if err != nil {
			return nil
		}
		results = append(results, structpb.NewStructValue(s))
	}
	return structpb.NewListValue(&structpb.ListValue{Values: results})
}
