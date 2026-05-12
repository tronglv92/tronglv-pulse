package slicesx

// Chunk splits a slice into multiple sub-slices of a specified size.
// The final chunk may be smaller if there are not enough remaining elements.
func Chunk[T any](items []T, size int) [][]T {
	if size <= 0 {
		return nil
	}
	var chunks [][]T
	for size < len(items) {
		chunks = append(chunks, items[:size:size]) // fix capacity to avoid sharing backing array
		items = items[size:]
	}
	return append(chunks, items)
}

// Filter returns a new slice containing only the elements that satisfy the given predicate.
func Filter[T any](s []T, predicate func(T) bool) []T {
	var result []T
	for _, v := range s {
		if predicate(v) {
			result = append(result, v)
		}
	}
	return result
}

// Unique returns a new slice with duplicate elements removed.
// It works only for comparable types.
func Unique[T comparable](input []T) []T {
	seen := make(map[T]struct{})
	var result []T
	for _, v := range input {
		if _, ok := seen[v]; !ok {
			seen[v] = struct{}{}
			result = append(result, v)
		}
	}
	return result
}

// HasAny reports whether any of the given `needles` exist in the `haystack` slice.
// It returns true as soon as it finds a match, making it efficient for large slices.
//
// Parameters:
//   - haystack: a slice of strings to search in (e.g., user roles, permissions, tags).
//   - needles: one or more string values to search for within the haystack.
//
// Returns:
//   - true if at least one needle is found in the haystack; false otherwise.
//
// Example:
//
//	roles := []string{"admin", "editor"}
//	HasAny(roles, "viewer", "editor") // returns true
func HasAny(haystack []string, needles ...string) bool {
	set := make(map[string]struct{}, len(haystack))
	for _, item := range haystack {
		set[item] = struct{}{}
	}
	for _, needle := range needles {
		if _, ok := set[needle]; ok {
			return true
		}
	}
	return false
}
