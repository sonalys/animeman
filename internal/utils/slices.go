package utils

func Map[T1, T2 any](in []T1, f func(T1) T2) []T2 {
	out := make([]T2, 0, len(in))
	for i := range in {
		out = append(out, f(in[i]))
	}
	return out
}

func Transform[T any](in []T, fns ...func(T) T) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		value := in[i]

		for _, fn := range fns {
			value = fn(value)
		}

		out = append(out, value)
	}
	return out
}

func Filter[T any](in []T, filters ...func(T) bool) []T {
	out := make([]T, 0, len(in))
outer:
	for i := range in {
		for _, filter := range filters {
			if !filter(in[i]) {
				continue outer
			}
		}
		out = append(out, in[i])
	}
	return out
}
