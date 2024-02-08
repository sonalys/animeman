package utils

func ConvertInterfaceList[T1, T2 any](from []T1) []T2 {
	out := make([]T2, 0, len(from))
	for i := range from {
		out = append(out, any(from[i]).(T2))
	}
	return out
}
