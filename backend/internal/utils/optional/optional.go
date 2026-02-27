package optional

type Value[T any] struct {
	value T
	isSet bool
}

func NewValue[T any](v T) Value[T] {
	return Value[T]{
		value: v,
		isSet: true,
	}
}

func (v Value[T]) Get() (T, bool) {
	return v.value, v.isSet
}

func (v Value[T]) Or(defaultValue T) T {
	if !v.isSet {
		return defaultValue
	}
	return v.value
}

func (v Value[T]) IsSet() bool {
	return v.isSet
}

func (v Value[T]) Value() T {
	return v.value
}
