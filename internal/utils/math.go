package utils

import "golang.org/x/exp/constraints"

func Min[T constraints.Ordered](values ...T) (min T) {
	if len(values) == 0 {
		return
	}
	min = values[0]
	for i := range values {
		if values[i] < min {
			min = values[i]
		}
	}
	return min
}

func Max[T constraints.Ordered](values ...T) (max T) {
	if len(values) == 0 {
		return
	}
	max = values[0]
	for i := range values {
		if values[i] > max {
			max = values[i]
		}
	}
	return max
}
