package errutils

// Must receives (value, error) and returns value.
// It will panic if error is not nil.
func Must[T any](value T, err error) T {
	if err != nil {
		panic(err)
	}
	return value
}
