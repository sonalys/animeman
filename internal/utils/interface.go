package utils

// ConvertInterfaceList converts a list of interface 1 to a list of interface 2.
// Be sure the 2 interfaces are interchangeable, otherwise it will panic.
func ConvertInterfaceList[T1, T2 any](from []T1) []T2 {
	out := make([]T2, 0, len(from))
	for i := range from {
		out = append(out, any(from[i]).(T2))
	}
	return out
}
