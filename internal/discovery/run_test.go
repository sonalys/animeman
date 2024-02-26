package discovery

import (
	"reflect"
	"testing"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/require"
)

func Test_filterEpisodes(t *testing.T) {
	type args struct {
		list      []parser.ParsedNyaa
		latestTag string
	}
	tests := []struct {
		name string
		args args
		want []parser.ParsedNyaa
	}{
		{
			name: "empty",
			args: args{},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "no tag",
			args: args{
				latestTag: "",
				list: []parser.ParsedNyaa{
					{SeasonEpisodeTag: "!Show3 S03E01"},
					{SeasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []parser.ParsedNyaa{
				{SeasonEpisodeTag: "!Show3 S03E01"},
				{SeasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: "!Show3 S03E01",
				list: []parser.ParsedNyaa{
					{SeasonEpisodeTag: "!Show3 S03E01"},
					{SeasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []parser.ParsedNyaa{
				{SeasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: "!Show3 S03",
				list: []parser.ParsedNyaa{
					{SeasonEpisodeTag: "!Show3 S03E01"},
					{SeasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "same tag",
			args: args{
				latestTag: "!Show3 S0302",
				list: []parser.ParsedNyaa{
					{SeasonEpisodeTag: "!Show3 S03E01"},
					{SeasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []parser.ParsedNyaa{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := episodeFilterNew(tt.args.list, tt.args.latestTag, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseAndSort([]nyaa.Entry{})
		require.Empty(t, got)
	})
	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}
		got := parseAndSort(input)
		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].SeasonEpisodeTag, got[i].SeasonEpisodeTag) <= 0)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := getDownloadableEntries([]nyaa.Entry{}, "", animelist.AiringStatusAiring)
		require.Empty(t, got)
	})
	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := getDownloadableEntries(input, "", animelist.AiringStatusAiring)
		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].SeasonEpisodeTag, got[i].SeasonEpisodeTag) <= 0)
		}
	})
	t.Run("airing: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := getDownloadableEntries(input, "Show3 S03E02", animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[:1]), got)
	})
	t.Run("airing: with repeated tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := getDownloadableEntries(input, "Show3 S03E01", animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[0:1]), got)
	})

	t.Run("airing: with latestTag and quality", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03 720p"},
			{Title: "Show3: S03E03 1080p"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := getDownloadableEntries(input, "Show3 S03E02", animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[1:2]), got)
	})
	t.Run("aired: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := getDownloadableEntries(input, "Show3 S03E02", animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[:1]), got)
	})
	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := getDownloadableEntries(input, "", animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[2:]), got)
	})
	t.Run("aired: with batch, different qualities", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03 1220x760"},
			{Title: "Show3: S03 1080p"},
		}
		got := getDownloadableEntries(input, "", animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[1:]), got)
	})
	t.Run("aired: with batch, with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := getDownloadableEntries(input, "Show3 S03E02", animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[:1]), got)
	})
	t.Run("same tag and quality, different seeders", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03", Seeders: 1},
			{Title: "Show3: S03E03", Seeders: 10},
			{Title: "Show3: S03"},
		}
		got := getDownloadableEntries(input, "Show3 S03E02", animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSort(input[1:2]), got)
	})
}
