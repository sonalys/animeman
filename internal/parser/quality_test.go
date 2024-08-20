package parser

import (
	"testing"
)

func Test_matchQuality(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  int
	}{
		{
			name:  "provider and dash with tags",
			title: "[Provider] provider and dash with tags - 02 (1080p) [hash].mkv",
			want:  1080,
		},
		{
			name:  "without parentheses",
			title: "show name S01 1080p WEBRip DD+ x265-EMBER",
			want:  1080,
		},
		{
			name:  "no quality",
			title: "no quality - 01.mkv",
			want:  -1,
		},
		{
			name:  "between composed tags",
			title: "[EMBER] show name (2010) (Season 1) [tag] [1080p HEVC 10 bits] (another tag)",
			want:  1080,
		},
		{name: "2x15", title: "Show name 2x15 1080x720", want: 720},
		{name: "S02E15", title: "Show name S02E15", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualityMatch(tt.title); got != tt.want {
				t.Errorf("qualityMatch() = got '%v', want '%v'", got, tt.want)
			}
		})
	}
}
