package utils

// PointerOrDefault returns value if it is not nil, otherwise it returns fallback.
func PointerOrDefault[T comparable](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}

// ValueOrDefault returns value if it is not the zero value, otherwise it returns fallback.
func ValueOrDefault[T comparable](value T, fallback T) T {
	if value == *new(T) {
		return fallback
	}
	return value
}
