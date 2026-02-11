package optional

// Coalesce returns value if not default, otherwise returns fallback.
func Coalesce[T comparable](value, fallback T) T {
	var empty T
	if value == empty {
		return fallback
	}
	return value
}
