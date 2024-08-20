package parser

import (
	"testing"
)

func Test_matchSeason(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{
			name:  "episode with provider and dash",
			title: "[Provider] episode with dash - 02 (1080p) [hash].mkv",
			want:  "1",
		},
		{
			name:  "episode with S01",
			title: "episode with S01 S01 1080p tag1 DD+ x265-EMBER",
			want:  "1",
		},
		{
			name:  "episode with dash",
			title: "episode with dash - 01.mkv",
			want:  "1",
		},
		{
			name:  "episode with source and tags between",
			title: "[EMBER] show name (2010) (Season 1) [tag] [1080p HEVC 10 bits] (another tag)",
			want:  "1",
		},
		{name: "2x15", title: "Showname 2x15", want: "2"},
		{name: "2 - 05", title: "Showname 2 - 05", want: "2"},
		{name: "S02E15", title: "Showname S02E15", want: "2"},
		{
			name:  "all dots",
			title: "Show.name.S02E19.subtitle.here.1080p.TAG.AAC2.0.H.264-VARYG.mkv",
			want:  "2",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := SeasonParse(tt.title); got != tt.want {
				t.Errorf("seasonMatch() = season '%v', want '%v'", got, tt.want)
			}
		})
	}
}
