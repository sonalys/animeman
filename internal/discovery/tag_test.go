package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/tags"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []torrentclient.Torrent
	}
	tests := []struct {
		name string
		args args
		want tags.Tag
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
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{1, 13},
			},
		},
		{
			name: "same season half episode",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"ore dake level up na ken", "S1E7"}},
					{Tags: []string{"solo leveling", "S1E7.5"}},
				},
			},
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{7.5},
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
			want: tags.Tag{
				Seasons: []int{3},
			},
		},
		{
			name: "empty",
			want: tags.Tag{},
		},
		{
			name: "one tag",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S01"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{1},
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
			want: tags.Tag{
				Seasons:  []int{1},
				Episodes: []float64{3},
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
			want: tags.Tag{
				Seasons:  []int{3},
				Episodes: []float64{1},
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
			want: tags.Tag{
				Seasons: []int{3},
			},
		},
		{
			name: "batches of different seasons",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S4"}},
				},
			},
			want: tags.Tag{
				Seasons: []int{4},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestTag(tt.args.torrents); got.Compare(tt.want) != 0 {
				t.Errorf("getLatestTag() = %v, want %v", got, tt.want)
			}
		})
	}
}
