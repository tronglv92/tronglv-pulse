package oncex

import "sync"

// OnceValue provides lazy, thread-safe initialization of a value.
// The initializer function is executed at most once.
type OnceValue[T any] struct {
	once sync.Once
	val  T
	err  error
}

// Get returns the initialized value.
// The init function is called only once, even under concurrent access.
func (o *OnceValue[T]) Get(init func() (T, error)) (T, error) {
	o.once.Do(func() {
		o.val, o.err = init()
	})
	return o.val, o.err
}

// MustGet returns the initialized value or panics on error.
// Useful for required dependencies at startup.
func (o *OnceValue[T]) MustGet(init func() T) T {
	o.once.Do(func() {
		o.val = init()
	})
	return o.val
}
