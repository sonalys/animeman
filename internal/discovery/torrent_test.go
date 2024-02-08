package discovery

import (
	"testing"

	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

func Test_getLatestTag(t *testing.T) {
	type args struct {
		torrents []torrentclient.Torrent
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "batch and same season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S3"}},
					{Tags: []string{"S3E2"}},
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
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S01"}},
				},
			},
			want: "S01",
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
			want: "S1E3",
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
			want: "S3E1",
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
			want: "S3",
		},
		{
			name: "batch and season",
			args: args{
				torrents: []torrentclient.Torrent{
					{Tags: []string{"S1E1~13"}},
					{Tags: []string{"S1E2"}},
					{Tags: []string{"S1E3"}},
				},
			},
			want: "S1E1~13",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := tagGetLatest(tt.args.torrents...); got != tt.want {
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
