package discovery

import (
	"reflect"
	"testing"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/require"
)

func Test_filterEpisodes(t *testing.T) {
	type args struct {
		list      []ParsedNyaa
		latestTag string
	}
	tests := []struct {
		name string
		args args
		want []ParsedNyaa
	}{
		{
			name: "empty",
			args: args{},
			want: []ParsedNyaa{},
		},
		{
			name: "no tag",
			args: args{
				latestTag: "",
				list: []ParsedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []ParsedNyaa{
				{seasonEpisodeTag: "!Show3 S03E01"},
				{seasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: "!Show3 S03E01",
				list: []ParsedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []ParsedNyaa{
				{seasonEpisodeTag: "!Show3 S03E02"},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: "!Show3 S03",
				list: []ParsedNyaa{
					{seasonEpisodeTag: "!Show3 S03E01"},
					{seasonEpisodeTag: "!Show3 S03E02"},
				},
			},
			want: []ParsedNyaa{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := episodeFilter(tt.args.list, tt.args.latestTag, false); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseNyaaEntries([]nyaa.Entry{})
		require.Empty(t, got)
	})
	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}
		got := parseNyaaEntries(input)
		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].seasonEpisodeTag, got[i].seasonEpisodeTag) <= 0)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := filterNyaaFeed([]nyaa.Entry{}, "", animelist.AiringStatusAiring)
		require.Empty(t, got)
	})
	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterNyaaFeed(input, "", animelist.AiringStatusAiring)

		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].seasonEpisodeTag, got[i].seasonEpisodeTag) <= 0)
		}
	})

	t.Run("airing: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterNyaaFeed(input, "Show3 S03E02", animelist.AiringStatusAiring)

		require.Len(t, got, 1)
		require.Equal(t, parseNyaaEntries(input[:1]), got)
	})

	t.Run("aired: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterNyaaFeed(input, "Show3 S03E02", animelist.AiringStatusAired)

		require.Len(t, got, 1)
		require.Equal(t, parseNyaaEntries(input[:1]), got)
	})

	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := filterNyaaFeed(input, "", animelist.AiringStatusAired)

		require.Len(t, got, 1)
		require.Equal(t, parseNyaaEntries(input[2:]), got)
	})

	t.Run("aired: with batch, with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := filterNyaaFeed(input, "Show3 S03E02", animelist.AiringStatusAired)

		require.Len(t, got, 1)
		require.Equal(t, parseNyaaEntries(input[:1]), got)
	})
}
