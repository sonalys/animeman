package utils

func ForEach[T, T1 any](in []T, f func(T) T1) []T1 {
	out := make([]T1, 0, len(in))
	for i := range in {
		out = append(out, f(in[i]))
	}
	return out
}

func Filter[T any](in []T, f func(T) bool) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		if f(in[i]) {
			out = append(out, in[i])
		}
	}
	return out
}
