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
			name: "codec tokens are extracted into params",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 AV1", "1080 HEVC", "1080"}},
			want: "1080",
		},
		{
			name: "no common tokens",
			opt:  torrentsource.SearchOptions{Qualities: []string{"720", "1080"}},
			want: "1080|720",
		},
		{
			name: "all tokens common",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 HEVC", "1080 HEVC"}},
			want: "1080",
		},
		{
			name: "multi-word quality",
			opt:  torrentsource.SearchOptions{Qualities: []string{"1080 H.265"}},
			want: "1080",
		},
		{
			name: "suffix",
			opt:  torrentsource.SearchOptions{SearchSuffix: `-"dub"`},
			want: `-"dub"`,
		},
		{
			// Sources are not part of `q` anymore, they are filtered from results.
			name: "qualities and suffix, sources ignored",
			opt: torrentsource.SearchOptions{
				Qualities:    []string{"1080 AV1", "1080 HEVC"},
				Sources:      []string{"Erai-raws", "SubsPlease"},
				SearchSuffix: `-"dub"`,
			},
			want: `1080 -"dub"`,
		},
		{
			// 1080 and 720 never co-occur, so the four qualities collapse
			// into one OR dimension; codecs go to params.
			name: "1080 and 720, either AV1 or HEVC",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 AV1", "1080 HEVC", "720 AV1", "720 HEVC"},
			},
			want: "1080|720",
		},
		{
			// "1080" alone is a subset of the others, so 1080 is common.
			name: "bare quality collapses into common tokens",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 AV1", "1080 HEVC", "1080"},
			},
			want: "1080",
		},
		{
			// "1080" alone is a subset of "1080 HEVC", so 1080 is common.
			name: "bare quality plus one codec",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 HEVC", "1080"},
			},
			want: "1080",
		},
		{
			// 720 never co-occurs with 1080, so resolutions form one dimension.
			name: "mixed bare and combined qualities",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 HEVC", "720 HEVC", "1080"},
			},
			want: "1080|720",
		},
		{
			// Video type tokens are extracted too, leaving only the resolution.
			name: "video type tokens",
			opt: torrentsource.SearchOptions{
				Qualities: []string{"1080 WEB-DL", "1080 BD"},
			},
			want: "1080",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, buildQuery(tt.opt))
		})
	}
}

func Test_splitQualityFilters(t *testing.T) {
	tests := []struct {
		name           string
		qualities      []string
		wantText       []string
		wantVideoType  string
		wantVideoCodec string
	}{
		{
			name:           "empty",
			qualities:      nil,
			wantText:       nil,
			wantVideoType:  "",
			wantVideoCodec: "",
		},
		{
			name:           "resolution only",
			qualities:      []string{"1080"},
			wantText:       []string{"1080"},
			wantVideoType:  "",
			wantVideoCodec: "",
		},
		{
			name:           "codec aliases deduplicate",
			qualities:      []string{"1080 HEVC", "1080 x265", "1080 H.265"},
			wantText:       []string{"1080", "1080", "1080"},
			wantVideoType:  "",
			wantVideoCodec: "2",
		},
		{
			name:           "multiple codecs",
			qualities:      []string{"1080 HEVC", "1080 AV1"},
			wantText:       []string{"1080", "1080"},
			wantVideoType:  "",
			wantVideoCodec: "2,3",
		},
		{
			name:           "video types",
			qualities:      []string{"1080 WEB-DL", "1080 BD"},
			wantText:       []string{"1080", "1080"},
			wantVideoType:  "11,9",
			wantVideoCodec: "",
		},
		{
			name:           "mixed codecs and types",
			qualities:      []string{"1080 WEB-DL HEVC", "1080 BD AVC"},
			wantText:       []string{"1080", "1080"},
			wantVideoType:  "11,9",
			wantVideoCodec: "1,2",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			text, params := splitQualityFilters(tt.qualities)
			require.Equal(t, tt.wantText, text)
			require.Equal(t, tt.wantVideoType, params.Get("video_type"))
			require.Equal(t, tt.wantVideoCodec, params.Get("video_codec"))
		})
	}
}
