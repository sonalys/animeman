package discovery

import (
	"testing"

	"github.com/expr-lang/expr"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/stretchr/testify/require"
)

func TestController_buildTorrentName(t *testing.T) {
	tests := []struct {
		name       string
		dep        Dependencies
		title      string
		parsedNyaa parser.ParsedNyaa
		want       string
	}{
		{
			name: "build torrent name with all placeholders",
			dep: Dependencies{
				Config: Config{
					RenameFormat: utils.Must(expr.Compile(`
						join(
							filter(
								[
									format("[%s]", releaseGroup), 
									title, 
									tag.LastEpisode() > 0 ? tag.String() : "", 
									format("[%dp]", verticalResolution),
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
			parsedNyaa: parser.ParsedNyaa{
				ExtractedMetadata: parser.Metadata{
					ReleaseGroup:       "release-group",
					Labels:             []string{"HEVC", "10bit"},
					Tag:                tags.Tag{Seasons: []int{1}, Episodes: []float64{1}},
					Title:              "My Anime Title",
					VerticalResolution: 1080,
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
