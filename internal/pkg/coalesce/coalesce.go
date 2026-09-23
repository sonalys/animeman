package coalesce

// Coalesce returns value if it is not nil, otherwise it returns fallback.
func Coalesce[T comparable](value *T, fallback T) T {
	if value == nil {
		return fallback
	}
	return *value
}

// OrDefault returns value if it is not the zero value, otherwise it returns fallback.
func OrDefault[T comparable](value T, fallback T) T {
	if value == *new(T) {
		return fallback
	}
	return value
}
