package discovery

import (
	"testing"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/stretchr/testify/require"
)

func Test_parseResults(t *testing.T) {
	tests := []struct {
		name    string
		entry   animelist.Entry
		results []torrentsource.Torrent
		config  Config
		want    []parser.TorrentMetadata
	}{
		{
			name: "honzuki S4",
			entry: animelist.Entry{
				Titles: []string{
					"Honzuki no Gekokujou: Shisho ni Naru Tame ni wa Shudan wo Erandeiraremasen 4th Season",
				},
			},
			results: []torrentsource.Torrent{
				{
					Title: "[Erai-raws] Honzuki no Gekokujou S4 - 21 [1080p CR WEB-DL AVC AAC][MultiSub][D12245DD]",
				},
			},
			want: []parser.TorrentMetadata{
				{
					Torrent: torrentsource.Torrent{
						Title: "[Erai-raws] Honzuki no Gekokujou S4 - 21 [1080p CR WEB-DL AVC AAC][MultiSub][D12245DD]",
					},
					Metadata: parser.Metadata{
						ReleaseGroup: "Erai-raws",
						Title:        "Honzuki no Gekokujou",
						Tag: tags.Tag{
							Seasons:  []int{4},
							Episodes: []float64{21},
						},
						Labels: []string{
							"1080p",
							"CR",
							"WEB-DL",
							"AVC",
							"AAC",
							"MultiSub",
							"D12245DD",
						},
						VerticalResolution: 1080,
					},
				},
			},
		},
		{
			name: "mushoku tensei III",
			entry: animelist.Entry{
				Titles: []string{"Mushoku Tensei III: Isekai Ittara Honki Dasu"},
			},
			results: []torrentsource.Torrent{
				{
					Title: "[Erai-raws] Mushoku Tensei III - Isekai Ittara Honki Dasu - 11 [1080p CR WEBRip HEVC AAC][MultiSub][6A0995D5]",
				},
			},
			want: []parser.TorrentMetadata{
				{
					Torrent: torrentsource.Torrent{
						Title: "[Erai-raws] Mushoku Tensei III - Isekai Ittara Honki Dasu - 11 [1080p CR WEBRip HEVC AAC][MultiSub][6A0995D5]",
					},
					Metadata: parser.Metadata{
						ReleaseGroup: "Erai-raws",
						Title:        "Mushoku Tensei III - Isekai Ittara Honki Dasu",
						Tag: tags.Tag{
							Seasons:  []int{1},
							Episodes: []float64{11},
						},
						Labels: []string{
							"1080p",
							"CR",
							"WEBRip",
							"HEVC",
							"AAC",
							"MultiSub",
							"6A0995D5",
						},
						VerticalResolution: 1080,
					},
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := parseResults(tt.entry, tt.results, tt.config)
			require.Equal(t, tt.want, got)
		})
	}
}
