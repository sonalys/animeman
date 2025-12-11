package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func Test_SeasonEpisodetag_BuildTag(t *testing.T) {
	testCases := []struct {
		name  string
		input SeasonEpisodeTag
		want  string
	}{
		{
			name:  "single season",
			input: SeasonEpisodeTag{Season: []int{1}},
			want:  "S1",
		},
		{
			name:  "multi season",
			input: SeasonEpisodeTag{Season: []int{1, 2}},
			want:  "S1-2",
		},
		{
			name:  "season single episode",
			input: SeasonEpisodeTag{Season: []int{1}, Episode: []float64{1}},
			want:  "S1E1",
		},
		{
			name:  "season multi episode",
			input: SeasonEpisodeTag{Season: []int{1}, Episode: []float64{1, 3.5}},
			want:  "S1E1-3.5",
		},
	}
	for _, tC := range testCases {
		t.Run(tC.name, func(t *testing.T) {
			got := tC.input.BuildTag()
			require.Equal(t, tC.want, got)
		})
	}
}
