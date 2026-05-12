package reflectx

import "reflect"

// IsZeroOfUnderlyingType checks if the given value v is the zero value for its type.
func IsZeroOfUnderlyingType(v interface{}) bool {
	return v == nil || reflect.DeepEqual(v, reflect.Zero(reflect.TypeOf(v)).Interface())
}
