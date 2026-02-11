package parser

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func Test_matchEpisode(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		episode []float64
		multi   bool
	}{
		{name: "S1E1~13", title: "S1E1~13", episode: []float64{1, 13}, multi: true},
		{name: "episode resolution", title: "S02E06.720p", episode: []float64{6}, multi: false},
		{name: "S1E7.5", title: "show name S1E7.5 ", episode: []float64{7.5}, multi: false},
		{name: "0x15", title: "show name 0x15", episode: []float64{15}, multi: false},
		{name: "-15", title: "show name - 15", episode: []float64{15}, multi: false},
		{name: "2 - 05", title: "show name 2 - 05", episode: []float64{5}, multi: false},
		{name: "S02E15", title: "show name S02E15", episode: []float64{15}, multi: false},
		{name: "Season", title: "show name Season 2", episode: []float64{}, multi: true},
		{name: "Season with episode", title: "show name Season 2 - 15", episode: []float64{15}, multi: false},
		{
			name:    "no episode",
			title:   "show name S01 1080p WEBRip DD+ x265-EMBER",
			episode: []float64{},
			multi:   true,
		},
		{
			name:    "x264 bug",
			title:   "show name S01E13 subtitle 1080p AAC2.0 H 264-VARYG",
			episode: []float64{13},
			multi:   false,
		},
		{
			name:    "all dots",
			title:   "show.name.S02E019.subtitle.here.1080p.AAC2.0.H.264-VARYG.mkv",
			episode: []float64{19},
			multi:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotEpisode := ParseEpisode(tt.title)
			assert.Equal(t, tt.episode, gotEpisode)
		})
	}
}
