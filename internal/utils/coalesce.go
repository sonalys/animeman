package utils

func Coalesce[T comparable](value, fallback T) T {
	var empty T
	if value == empty {
		return fallback
	}
	return value
}
