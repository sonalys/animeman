package discovery

import (
	"testing"

	"github.com/sonalys/animeman/integrations/qbittorrent"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []qbittorrent.Torrent
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "batch and same season",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S3"},
					{Tags: "S3E2"},
				},
			},
			want: "S3",
		},
		{
			name: "empty",
			want: "",
		},
		{
			name: "one tag",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S01"},
				},
			},
			want: "S01",
		},
		{
			name: "same season",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S1E1"},
					{Tags: "S1E2"},
					{Tags: "S1E3"},
				},
			},
			want: "S1E3",
		},
		{
			name: "different seasons",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S3E1"},
					{Tags: "S2E2"},
					{Tags: "S1E3"},
				},
			},
			want: "S3E1",
		},
		{
			name: "batch and season",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S3"},
					{Tags: "S2E2"},
					{Tags: "S1E3"},
				},
			},
			want: "S3",
		},
		{
			name: "batch and season",
			args: args{
				torrents: []qbittorrent.Torrent{
					{Tags: "S1E1~13"},
					{Tags: "S1E2"},
					{Tags: "S1E3"},
				},
			},
			want: "S1E1~13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := getLatestTag(tt.args.torrents...); got != tt.want {
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
			if got := mergeBatchEpisodes(tt.args.tag); got != tt.want {
				t.Errorf("mergeBatchEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}
