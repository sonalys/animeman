package tags_test

import (
	"testing"

	"github.com/sonalys/animeman/internal/pkg/tags"
	"github.com/stretchr/testify/require"
)

func Test_tagCompare(t *testing.T) {
	t.Run("same tag", func(t *testing.T) {
		tag := tags.Tag{
			Seasons:  []int{3},
			Episodes: []float64{2},
		}
		require.Zero(t, tag.Compare(tag))
	})

	t.Run("first episode and zero tag", func(t *testing.T) {
		tag := tags.Tag{
			Seasons:  []int{1},
			Episodes: []float64{1},
		}
		require.Greater(t, tag.Compare(tags.Tag{}), 0)
	})

	t.Run("batch different season", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{2},
		}

		tagB := tags.Tag{
			Seasons: []int{3},
		}

		require.Equal(t, tagA.Compare(tagB), -1)
	})

	t.Run("batch and single", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{2},
		}

		tagB := tags.Tag{
			Seasons:  []int{2},
			Episodes: []float64{1},
		}

		require.Equal(t, tagA.Compare(tagB), 1)
	})

	t.Run("batch and zero", func(t *testing.T) {
		tagA := tags.Tag{
			Seasons: []int{1},
		}

		require.Equal(t, tagA.Compare(tags.Zero), 1)
	})
}
