package parser

import (
	"testing"

	"github.com/sonalys/animeman/internal/utils/tags"
	"github.com/stretchr/testify/require"
)

func Test_SeasonEpisodetag_BuildTag(t *testing.T) {
	testCases := []struct {
		name  string
		input tags.Tag
		want  string
	}{
		{
			name:  "single season",
			input: tags.Tag{Seasons: []int{1}},
			want:  "S1",
		},
		{
			name:  "multi season",
			input: tags.Tag{Seasons: []int{1, 2}},
			want:  "S1-2",
		},
		{
			name:  "season single episode",
			input: tags.Tag{Seasons: []int{1}, Episodes: []float64{1}},
			want:  "S1E1",
		},
		{
			name:  "season multi episode",
			input: tags.Tag{Seasons: []int{1}, Episodes: []float64{1, 3.5}},
			want:  "S1E1-3.5",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := tC.input.String()
			require.Equal(t, tC.want, got)
		})
	}
}
