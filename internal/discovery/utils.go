package discovery

import (
	"strconv"

	"github.com/sonalys/animeman/internal/utils"
	"golang.org/x/exp/constraints"
)

func strSliceToFloat(from []string) []float64 {
	out := make([]float64, 0, len(from))
	for _, cur := range from {
		out = append(out, utils.Must(strconv.ParseFloat(cur, 64)))
	}
	return out
}

func min[T constraints.Ordered](values ...T) (min T) {
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

func max[T constraints.Ordered](values ...T) (max T) {
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
