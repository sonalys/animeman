package nekobt

import (
	"testing"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/stretchr/testify/require"
)

func Test_buildQuery(t *testing.T) {
	tests := []struct {
		name string
		opt  torrentsource.SearchOptions
		want string
	}{
		{
			name: "empty",
			opt:  torrentsource.SearchOptions{},
			want: "",
		},
		{
			name: "single quality",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080"}},
			want: "1080",
		},
		{
			name: "common tokens hoisted, differing tokens ORed",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 AV1", "1080 HEVC", "1080"}},
			want: "1080 AV1|HEVC",
		},
		{
			name: "no common tokens",
			opt:  torrentsource.SearchOptions{Qualities: []string{"720", "1080"}},
			want: "1080|720",
		},
		{
			name: "all tokens common",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 HEVC", "1080 HEVC"}},
			want: "1080 HEVC",
		},
		{
			name: "multi-word quality",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 H.265"}},
			want: "1080 H.265",
		},
		{
			name: "sources ORed",
			opt:  torrentsource.SearchOptions{Sources: []string{"Erai-raws", "SubsPlease"}},
			want: "Erai-raws|SubsPlease",
		},
		{
			name: "suffix",
			opt:  torrentsource.SearchOptions{SearchSuffix: `-"dub"`},
			want: `-"dub"`,
		},
		{
			name: "qualities, sources and suffix",
			opt: torrentsource.SearchOptions{
				Qualities:    []string{"1080 AV1", "1080 HEVC"},
				Sources:      []string{"Erai-raws", "SubsPlease"},
				SearchSuffix: `-"dub"`,
			},
			want: `1080 AV1|HEVC Erai-raws|SubsPlease -"dub"`,
		},
		{
			// 1080 and 720 never co-occur, same for AV1 and HEVC, so the four
			// qualities collapse into two OR dimensions.
			name: "1080 and 720, either AV1 or HEVC",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 AV1", "1080 HEVC", "720 AV1", "720 HEVC"},
			},
			want: "1080|720 AV1|HEVC",
		},
		{
			// "1080" alone is a subset of the others, so 1080 is common and
			// only HEVC/AV1 differ.
			name: "bare quality collapses into common tokens",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 AV1", "1080 HEVC", "1080"},
			},
			want: "1080 AV1|HEVC",
		},
		{
			// "1080" alone is a subset of "1080 HEVC", so 1080 is common and
			// HEVC is a single-token dimension.
			name: "bare quality plus one codec",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 HEVC", "1080"},
			},
			want: "1080 HEVC",
		},
		{
			// 720 never co-occurs with 1080, but HEVC co-occurs with both, so
			// HEVC is common and resolutions form one dimension.
			name: "mixed bare and combined qualities",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 HEVC", "720 HEVC", "1080"},
			},
			want: "1080|720 HEVC",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, buildQuery(tt.opt))
		})
	}
}
