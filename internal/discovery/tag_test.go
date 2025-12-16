package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
	"github.com/stretchr/testify/require"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []torrentclient.Torrent
	}
	tests := []struct {
		name string
		args args
		want parser.SeasonEpisodeTag
	}{
		{
			name: "batch and season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S1E1~13"}},
					{Tags: []string{"S1E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season:  []int{1},
				Episode: []float64{1, 13},
			},
		},
		{
			name: "same season half episode",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"ore dake level up na ken S1E7"}},
					{Tags: []string{"solo leveling S1E7.5"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season:  []int{1},
				Episode: []float64{7.5},
			},
		},
		{
			name: "batch and same season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S3E2"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season: []int{3},
			},
		},
		{
			name: "empty",
			want: parser.SeasonEpisodeTag{},
		},
		{
			name: "one tag",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S01"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season: []int{1},
			},
		},
		{
			name: "same season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S1E1"}},
					{Tags: []string{"S1E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season:  []int{1},
				Episode: []float64{3},
			},
		},
		{
			name: "different seasons",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S3E1"}},
					{Tags: []string{"S2E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season:  []int{3},
				Episode: []float64{1},
			},
		},
		{
			name: "batch and season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S2E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: parser.SeasonEpisodeTag{
				Season: []int{3},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestTag(tt.args.torrents); tagCompare(got, tt.want) != 0 {
				t.Errorf("getLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_mergeBatchEpisodes(t *testing.T) {
	type args struct {
		tag string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "empty",
			args: args{},
			want: "",
		},
		{
			name: "ok",
			args: args{
				tag: "S03E1~12",
			},
			want: "S03E12",
		},
		{
			name: "no batch",
			args: args{
				tag: "S03E1",
			},
			want: "S03E1",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagMergeBatchEpisodes(tt.args.tag); got != tt.want {
				t.Errorf("mergeBatchEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_tagCompare(t *testing.T) {
	t.Run("same tag", func(t *testing.T) {
		tag := parser.SeasonEpisodeTag{
			Season:  []int{3},
			Episode: []float64{2},
		}
		require.Zero(t, tagCompare(tag, tag))
	})
}
