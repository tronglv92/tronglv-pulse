package interfacex

import "fmt"

// ToInterfaceSlice converts a slice of concrete type T into a slice of interface type I.
// It silently returns nil if any of the items in the input slice do not implement the interface I.
//
// This function is useful when you want to treat a slice of structs or pointers to structs
// as a slice of interfaces they implement.
//
// For a version that returns an error instead of silently failing, use ToInterfaceSliceSafe.
func ToInterfaceSlice[T any, I any](input []T) []I {
	out, err := ToInterfaceSliceSafe[T, I](input)
	if err != nil {
		return nil
	}
	return out
}

// ToInterfaceSliceSafe safely converts a slice of concrete type T into a slice of interface type I.
// If any item in the slice does not implement the interface I, it returns an error.
//
// Example:
//
//	type MyStruct struct{}
//	func (MyStruct) Foo() {}
//	var items []MyStruct
//	out, err := ToInterfaceSliceSafe[MyStruct, MyInterface](items)
//
// This function is safer than ToInterfaceSlice because it provides error feedback
// when a type assertion fails.
func ToInterfaceSliceSafe[T any, I any](input []T) ([]I, error) {
	out := make([]I, len(input))
	for i, v := range input {
		iv, ok := any(v).(I)
		if !ok {
			return nil, fmt.Errorf("type %T does not implement interface", v)
		}
		out[i] = iv
	}
	return out, nil
}
