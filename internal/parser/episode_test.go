package parser

import "testing"

func Test_matchEpisode(t *testing.T) {
	tests := []struct {
		name    string
		title   string
		episode string
		multi   bool
	}{
		{name: "S1E1~13", title: "S1E1~13", episode: "1~13", multi: true},
		{name: "episode resolution", title: "S02E06.720p", episode: "6", multi: false},
		{name: "S1E7.5", title: "solo leveling S1E7.5 ", episode: "7.5", multi: false},
		{name: "0x15", title: "Frieren 0x15", episode: "15", multi: false},
		{name: "-15", title: "Frieren - 15", episode: "15", multi: false},
		{name: "2 - 05", title: "Frieren 2 - 05", episode: "5", multi: false},
		{name: "S02E15", title: "Frieren S02E15", episode: "15", multi: false},
		{name: "Season", title: "Frieren Season 2", episode: "", multi: true},
		{name: "Season with episode", title: "Frieren Season 2 - 15", episode: "15", multi: false},
		{
			name:    "Boku no Kokoro no Yabai",
			title:   "Boku no Kokoro no Yabai Yatsu S01 1080p WEBRip DD+ x265-EMBER",
			episode: "",
			multi:   true,
		},
		{
			name:    "264 bug",
			title:   "Undead Unluck S01E13 Tatiana 1080p HULU WEB-DL AAC2.0 H 264-VARYG",
			episode: "13",
			multi:   false,
		},
		{
			name:    "all dots",
			title:   "MASHLE.MAGIC.AND.MUSCLES.S02E19.Mash.Burnedead.and.the.Magical.Maestro.1080p.CR.WEB-DL.AAC2.0.H.264-VARYG.mkv",
			episode: "19",
			multi:   false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			episode, isMulti := EpisodeParse(tt.title)
			if episode != tt.episode {
				t.Errorf("episodeMatch() got episode = %v, want %v", episode, tt.episode)
			}
			if isMulti != tt.multi {
				t.Errorf("episodeMatch() got multi = %v, want %v", isMulti, tt.multi)
			}
		})
	}
}
