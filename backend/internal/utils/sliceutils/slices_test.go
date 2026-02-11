package sliceutils_test

import (
	"strconv"
	"testing"

	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"github.com/stretchr/testify/require"
)

func Test_Map(t *testing.T) {
	t.Run("should run and return new slice of type with value of func", func(t *testing.T) {
		from := []int{1, 2, 3}

		var to []string = sliceutils.Map(from, func(cur int) string {
			return strconv.Itoa(cur)
		})

		require.Equal(t, []string{"1", "2", "3"}, to)
	})
}

func Test_Filter(t *testing.T) {
	t.Run("should filter with and condition for all filters", func(t *testing.T) {
		from := []int{1, 2, 3}

		got := sliceutils.Filter(from,
			func(cur int) bool { return cur == 2 },
			func(cur int) bool { return true },
		)

		require.Equal(t, []int{2}, got)
	})
}
