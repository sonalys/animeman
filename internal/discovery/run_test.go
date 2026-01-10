package discovery

import (
	"reflect"
	"testing"
	"time"

	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/require"
)

func Test_filterEpisodes(t *testing.T) {
	type args struct {
		list      []parser.ParsedNyaa
		latestTag tags.Tag
	}
	tests := []struct {
		name string
		args args
		want []parser.ParsedNyaa
	}{
		{
			name: "s\\de\\d{2}",
			args: args{
				latestTag: tags.SeasonEpisode(1, 18),
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(1, 16)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(1, 17)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(1, 18)}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "empty",
			args: args{},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "no tag",
			args: args{
				latestTag: tags.Zero,
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{
				{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
				{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: tags.SeasonEpisode(3, 1),
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{
				{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: tags.Tag{Seasons: []int{3}},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "same tag",
			args: args{
				latestTag: tags.SeasonEpisode(3, 2),
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{Meta: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterNewEpisodes(tt.args.list, tt.args.latestTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseResults([]nyaa.Entry{})
		got = sortResults(animelist.Entry{}, got)
		require.Empty(t, got)
	})
	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}

		got := parseResults(input)
		got = sortResults(animelist.Entry{}, got)

		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].Meta.Tag, got[i].Meta.Tag) <= 0)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	newEntry := func(airingStatus animelist.AiringStatus) animelist.Entry {
		return animelist.NewEntry(nil, animelist.ListStatusWatching, airingStatus, time.Now(), 0)
	}

	t.Run("empty", func(t *testing.T) {
		got := filterRelevantResults(animelist.Entry{}, []parser.ParsedNyaa{}, tags.Zero)
		require.Empty(t, got)
	})

	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(animelist.Entry{}, parsed, tags.Zero)

		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].Meta.Tag, got[i].Meta.Tag) <= 0)
		}
	})

	t.Run("airing: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		entry := newEntry(animelist.AiringStatusAiring)
		parsedTorrents := parseResults(input)
		latestTag := tags.SeasonEpisode(3, 2)

		got := filterRelevantResults(entry, parsedTorrents, latestTag)

		require.Equal(t, parsedTorrents[:1], got)
	})

	t.Run("airing: with repeated tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(animelist.Entry{}, parsed, tags.SeasonEpisode(3, 1))

		require.Len(t, got, 1)
		require.Equal(t, parsed[0:1], got)
	})

	t.Run("airing: with latestTag and quality", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03 720p"},
			{Title: "Show3: S03E03 1080p"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(animelist.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:2], got)
	})

	t.Run("aired: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(animelist.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[:1], got)
	})

	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(animelist.AiringStatusAired), parsed, tags.Zero)

		require.Len(t, got, 1)
		require.Equal(t, parsed[2:], got)
	})

	t.Run("aired: with batch, different qualities", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03 1220x760"},
			{Title: "Show3: S03 1080p"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(animelist.AiringStatusAired), parsed, tags.Zero)

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:], got)
	})

	t.Run("aired: with batch, with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(animelist.AiringStatusAired), parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[:1], got)
	})

	t.Run("same tag and quality, different seeders", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03", Seeders: 1},
			{Title: "Show3: S03E03", Seeders: 10},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(animelist.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:2], got)
	})
}

func Test_calculateTitleSimilarityScore(t *testing.T) {
	t.Run("exact match in lower case", func(t *testing.T) {
		score := calculateTitleSimilarityScore("My pony academy: the story continues", "My Pony Academy the story continues")
		require.EqualValues(t, score, 1)
	})

	t.Run("closer match should have higher score", func(t *testing.T) {
		originalTitle := "My pony academy: the battle continues"
		scoreA := calculateTitleSimilarityScore(originalTitle, "My Pony Academy")
		scoreB := calculateTitleSimilarityScore(originalTitle, "My Pony Academy 2: second battle")
		require.Greater(t, scoreA, scoreB)
	})
}
