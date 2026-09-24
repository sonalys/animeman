package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/stretchr/testify/require"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []torrentclient.Torrent
	}
	tests := []struct {
		name string
		args args
		want metadata.Tag
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
			want: metadata.Tag{
				Number:   1,
				Episodes: []metadata.EpisodeRange{{Start: 1, End: 13}},
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
			want: metadata.Tag{
				Number:   1,
				Episodes: []metadata.EpisodeRange{{Start: 7.5}},
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
			want: metadata.Tag{
				Number: 3,
			},
		},
		{
			name: "empty",
			want: metadata.Tag{},
		},
		{
			name: "one tag",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S01"}},
				},
			},
			want: metadata.Tag{
				Number: 1,
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
			want: metadata.Tag{
				Number:   1,
				Episodes: []metadata.EpisodeRange{{Start: 3}},
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
			want: metadata.Tag{
				Number:   3,
				Episodes: []metadata.EpisodeRange{{Start: 1}},
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
			want: metadata.Tag{
				Number: 3,
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
			want: metadata.Tag{
				Number: 4,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := getLatestTag(tt.args.torrents)
			require.Equal(t, tt.want, got)
		})
	}
}
