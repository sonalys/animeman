package parser

import (
	"testing"

	"github.com/sonalys/animeman/internal/pkg/tags"
	"github.com/stretchr/testify/require"
)

func TestTitleParse(t *testing.T) {
	tests := []struct {
		name  string
		title string
		want  Metadata
	}{
		{
			name:  "simple",
			title: "[Erai-raws] Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen - 07 [1080p CR WEB-DL AVC AAC][MultiSub][D834BF79]",
			want: Metadata{
				Title: "Jujutsu Kaisen: Shimetsu Kaiyuu - Zenpen",
				Tag: tags.Tag{
					Seasons:  []int{1},
					Episodes: []float64{7},
				},
				VerticalResolution: 1080,
				ReleaseGroup:       "Erai-raws",
				Labels: []string{
					"1080p",
					"CR",
					"WEB-DL",
					"AVC",
					"AAC",
					"MultiSub",
					"D834BF79",
				},
			},
		},
		{
			name:  "simple",
			title: "[Provider] Show name - 07 (1080p) [file-hash]",
			want: Metadata{
				Title: "Show name",
				Tag: tags.Tag{
					Seasons:  []int{1},
					Episodes: []float64{7},
				},
				VerticalResolution: 1080,
				ReleaseGroup:       "Provider",
				Labels:             []string{"file-hash"},
			},
		},
		{
			name:  "half episode",
			title: "[Provider] Show name - 07.5 (1080p) [9F8A2A07].mkv",
			want: Metadata{
				Title: "Show name",
				Tag: tags.Tag{
					Seasons:  []int{1},
					Episodes: []float64{7.5},
				},
				VerticalResolution: 1080,
				ReleaseGroup:       "Provider",
				Labels:             []string{"9F8A2A07"},
			},
		},
		{
			name:  "half episode at the end of string",
			title: "[Provider] Show name - 07.5",
			want: Metadata{
				Title: "Show name",
				Tag: tags.Tag{
					Seasons:  []int{1},
					Episodes: []float64{7.5},
				},
				VerticalResolution: -1,
				ReleaseGroup:       "Provider",
				Labels:             []string{},
			},
		},
		{
			name:  "all dots",
			title: "Show.name.S02E19.subtitle.here.1080p.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			want: Metadata{
				Title: "Show name",
				Tag: tags.Tag{
					Seasons:  []int{2},
					Episodes: []float64{19},
				},
				VerticalResolution: 1080,
			},
		},
		{
			name:  "all dots 2",
			title: "Show.name.S01E20.Anno.Un.1080p.HULU.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			want: Metadata{
				Title: "Show name",
				Tag: tags.Tag{
					Seasons:  []int{1},
					Episodes: []float64{20},
				},
				VerticalResolution: 1080,
			},
		},
		{
			name:  "lowercase",
			title: "show s02e02",
			want: Metadata{
				Title: "show",
				Tag: tags.Tag{
					Seasons:  []int{2},
					Episodes: []float64{2},
				},
				VerticalResolution: -1,
			},
		},
		{
			name:  "curly braces tags suffix",
			title: "[Erai-raws] Honzuki no Gekokujou S4 - 21 [1080p CR WEB-DL AVC AAC][MultiSub][D12245DD] {Tags:L0;V9;C1;A=ja;S=en,ptbr,es419,eses,ar,frfr,de,it,ru,id,ms,th,vi,zhhans,zhhant,pl;}",
			want: Metadata{
				Title: "Honzuki no Gekokujou",
				Tag: tags.Tag{
					Seasons:  []int{4},
					Episodes: []float64{21},
				},
				VerticalResolution: 1080,
				ReleaseGroup:       "Erai-raws",
				Labels:             []string{"1080p", "CR", "WEB-DL", "AVC", "AAC", "MultiSub", "D12245DD"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := Parse(tt.title, 1, nil)
			require.Equal(t, tt.want, got)
		})
	}
}
