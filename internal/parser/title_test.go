package parser

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestTitleStrip(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  string
	}{
		{name: "\"quoted\" title", title: "\"quoted\" title", want: "quoted title"},
		{name: "title.with.dots", title: "title.with.dots", want: "title.with.dots"},
		{name: "empty", title: "", want: ""},
		{name: "multiple spaces", title: "My     cool   anime", want: "My cool anime"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleStrip(tt.title, false); got != tt.want {
				t.Errorf("TitleStrip() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestTitleParse(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  Metadata
	}{
		{
			name:  "lv2",
			title: "[SubsPlease] Boku no Hero Academia - 140 (1080p) [CAE71930].mkv",
			want: Metadata{
				Title:              "Boku no Hero Academia",
				Episode:            "140",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
				Source:             "SubsPlease",
				Tags:               []string{"CAE71930"},
			},
		},
		{
			name:  "lv2",
			title: "[SubsPlease] Lv2 kara Cheat datta Motoyuusha Kouho no Mattari Isekai Life - 07 (1080p) [5E653DF8]",
			want: Metadata{
				Title:              "Lv2 kara Cheat datta Motoyuusha Kouho no Mattari Isekai Life",
				Episode:            "7",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
				Source:             "SubsPlease",
				Tags:               []string{"5E653DF8"},
			},
		},
		{
			name:  "half episode",
			title: "[SubsPlease] Solo Leveling - 07.5 (1080p) [9F8A2A07].mkv",
			want: Metadata{
				Title:              "Solo Leveling",
				Episode:            "7.5",
				Season:             "1",
				VerticalResolution: 1080,
				Source:             "SubsPlease",
				Tags:               []string{"9F8A2A07"},
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "half episode at the end of string",
			title: "[SubsPlease] Solo Leveling - 07.5",
			want: Metadata{
				Title:              "Solo Leveling",
				Episode:            "7.5",
				Season:             "1",
				VerticalResolution: -1,
				Source:             "SubsPlease",
				Tags:               []string{},
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "all dots",
			title: "MASHLE.MAGIC.AND.MUSCLES.S02E19.Mash.Burnedead.and.the.Magical.Maestro.1080p.CR.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			want: Metadata{
				Title:              "MASHLE MAGIC AND MUSCLES",
				Episode:            "19",
				Season:             "2",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "undead unluck",
			title: "Undead.Unluck.S01E20.Anno.Un.1080p.HULU.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			want: Metadata{
				Title:              "Undead Unluck",
				Episode:            "20",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.title)
			require.Equal(t, tt.want, got)
		})
	}
}
