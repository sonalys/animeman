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
		opts  []TitleStripOptions
		want  Metadata
	}{
		{
			name:  "simple",
			title: "[Provider] Show name - 140 (1080p) [file-hash].mkv",
			want: Metadata{
				Title:              "Show name",
				Episode:            "140",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
				Source:             "Provider",
				Tags:               []string{"file-hash"},
			},
		},
		{
			name:  "simple",
			title: "[Provider] Show name - 07 (1080p) [file-hash]",
			want: Metadata{
				Title:              "Show name",
				Episode:            "7",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
				Source:             "Provider",
				Tags:               []string{"file-hash"},
			},
		},
		{
			name:  "half episode",
			title: "[Provider] Show name - 07.5 (1080p) [9F8A2A07].mkv",
			want: Metadata{
				Title:              "Show name",
				Episode:            "7.5",
				Season:             "1",
				VerticalResolution: 1080,
				Source:             "Provider",
				Tags:               []string{"9F8A2A07"},
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "half episode at the end of string",
			title: "[Provider] Show name - 07.5",
			want: Metadata{
				Title:              "Show name",
				Episode:            "7.5",
				Season:             "1",
				VerticalResolution: -1,
				Source:             "Provider",
				Tags:               []string{},
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "all dots",
			title: "Show.name.S02E19.subtitle.here.1080p.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			opts:  []TitleStripOptions{RemoveDots()},
			want: Metadata{
				Title:              "Show name",
				Episode:            "19",
				Season:             "2",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
			},
		},
		{
			name:  "all dots 2",
			title: "Show.name.S01E20.Anno.Un.1080p.HULU.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			opts:  []TitleStripOptions{RemoveDots()},
			want: Metadata{
				Title:              "Show name",
				Episode:            "20",
				Season:             "1",
				VerticalResolution: 1080,
				IsMultiEpisode:     false,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.title, tt.opts...)
			require.Equal(t, tt.want, got)
		})
	}
}
