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
		{name: "empty", title: "", want: ""},
		{name: "multiple spaces", title: "My     cool   anime", want: "My cool anime"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := TitleStrip(tt.title); got != tt.want {
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
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := TitleParse(tt.title)
			require.Equal(t, tt.want, got)
		})
	}
}
