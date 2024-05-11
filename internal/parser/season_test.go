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
			name:  "Ragna Crimson",
			title: "[SubsPlease] Ragna Crimson - 02 (1080p) [B8FB702D].mkv",
			want:  "1",
		},
		{
			name:  "Boku no Kokoro no Yabai",
			title: "Boku no Kokoro no Yabai Yatsu S01 1080p WEBRip DD+ x265-EMBER",
			want:  "1",
		},
		{
			name:  "Kusuriya no Hitorigoto",
			title: "Kusuriya no Hitorigoto - 01.mkv",
			want:  "1",
		},
		{
			name:  "Tatami Galaxy",
			title: "[EMBER] The Tatami Galaxy (2010) (Season 1) [BDRip] [1080p HEVC 10 bits] (Yojouhan Shinwa Taikei)",
			want:  "1",
		},
		{name: "2x15", title: "Frieren 2x15", want: "2"},
		{name: "2 - 05", title: "Frieren 2 - 05", want: "2"},
		{name: "S02E15", title: "Frieren S02E15", want: "2"},
		{
			name:  "all dots",
			title: "MASHLE.MAGIC.AND.MUSCLES.S02E19.Mash.Burnedead.and.the.Magical.Maestro.1080p.CR.WEB-DL.AAC2.0.H.264-VARYG.mkv",
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
