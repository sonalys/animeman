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
			name:  "Ragna Crimson",
			title: "[SubsPlease] Ragna Crimson - 02 (1080p) [B8FB702D].mkv",
			want:  1080,
		},
		{
			name:  "Boku no Kokoro no Yabai",
			title: "Boku no Kokoro no Yabai Yatsu S01 1080p WEBRip DD+ x265-EMBER",
			want:  1080,
		},
		{
			name:  "Kusuriya no Hitorigoto",
			title: "Kusuriya no Hitorigoto - 01.mkv",
			want:  -1,
		},
		{
			name:  "Tatami Galaxy",
			title: "[EMBER] The Tatami Galaxy (2010) (Season 1) [BDRip] [1080p HEVC 10 bits] (Yojouhan Shinwa Taikei)",
			want:  1080,
		},
		{name: "2x15", title: "Frieren 2x15 1080x720", want: 720},
		{name: "S02E15", title: "Frieren S02E15", want: -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := qualityMatch(tt.title); got != tt.want {
				t.Errorf("qualityMatch() = got '%v', want '%v'", got, tt.want)
			}
		})
	}
}
