package controller

import (
	"testing"
	"time"

	"github.com/expr-lang/expr"
	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/stretchr/testify/require"
)

func TestController_buildTorrentName(t *testing.T) {
	tests := []struct {
		name       string
		dep        Dependencies
		title      string
		parsedNyaa torrentsource.Torrent
		want       string
	}{
		{
			name: "build torrent name with all placeholders",
			dep: Dependencies{
				Config: Config{
					RenameFormat: must.Must(expr.Compile(`
						join(
							filter(
								[
									format("[%s]", releaseGroup), 
									title, 
									tag.IsZero() ? "" : tag.String(), 
									format("[%s]", verticalResolution),
									format("%v", map(labels, upper(#))),
								], 
								# != "",
							), 
							" ",
						)
			`)),
				},
			},
			title: "My Anime Title",
			parsedNyaa: torrentsource.Torrent{
				Metadata: metadata.Metadata{
					Group:  "release-group",
					Labels: []string{"HEVC", "10bit"},
					Tags: []metadata.Tag{
						{Number: 1, Episodes: []metadata.EpisodeRange{{Start: 1}}},
					},
					PrimaryTitle: "My Anime Title",
					Resolutions:  []string{"1080p"},
				},
			},
			want: "[release-group] My Anime Title S1E1 [1080p] [HEVC 10BIT]",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := New(tt.dep)
			got := c.buildTorrentName(tt.title, tt.parsedNyaa)
			require.Equal(t, tt.want, got)
		})
	}
}

func Test_normalizeTitle(t *testing.T) {
	t0 := time.Time{}

	entries := []animelist.Entry{
		animelist.NewEntry(
			[]string{"Sousou no Frieren", "Frieren: Beyond Journey's End"},
			animelist.ListStatusWatching,
			animelist.AiringStatusAired,
			t0,
			t0,
			28,
			nil,
		),
		animelist.NewEntry(
			[]string{"Ore dake Level Up na Ken"},
			animelist.ListStatusWatching,
			animelist.AiringStatusAiring,
			t0,
			t0,
			12,
			nil,
		),
	}

	t.Run("alternative title gets normalized", func(t *testing.T) {
		// Torrent named after an alternative title with hyphen and subtitle.
		got := normalizeTitle("Ore dake Level-Up na Ken: Season 2", entries)
		require.Equal(t, "Ore dake Level Up na Ken", got)
	})

	t.Run("no match returns original title", func(t *testing.T) {
		got := normalizeTitle("Completely Unrelated Show", entries)
		require.Equal(t, "Completely Unrelated Show", got)
	})

	t.Run("empty entries returns original title", func(t *testing.T) {
		got := normalizeTitle("Some Show", nil)
		require.Equal(t, "Some Show", got)
	})
}

// Do not break compatibility, or the torrent client will re-download everything again without proper discovery.
func Test_buildTitleTag(t *testing.T) {
	cases := []struct {
		name  string
		input string
		want  string
	}{
		{
			name:  "alpha-numerical only",
			input: "Honzuki no Gekokujou: Ryoushu no Youjo",
			want:  "!honzuki no gekokujou ryoushu no youjo",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := buildTitleTag(tc.input)
			require.Equal(t, tc.want, got)
		})
	}
}
