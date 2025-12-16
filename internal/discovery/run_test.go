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
		latestTag parser.SeasonEpisodeTag
	}
	tests := []struct {
		name string
		args args
		want []parser.ParsedNyaa
	}{
		{
			name: "s\\de\\d{2}",
			args: args{
				latestTag: parser.SeasonEpisodeTag{
					Season:  []int{1},
					Episode: []float64{18},
				},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{1}, Episode: []float64{16}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{1}, Episode: []float64{17}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{1}, Episode: []float64{18}}}},
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
				latestTag: parser.SeasonEpisodeTag{},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
				},
			},
			want: []parser.ParsedNyaa{
				{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}}},
				{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
				},
			},
			want: []parser.ParsedNyaa{
				{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: parser.SeasonEpisodeTag{Season: []int{3}},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "same tag",
			args: args{
				latestTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}},
				list: []parser.ParsedNyaa{
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}}},
					{Meta: parser.Metadata{SeasonEpisodeTag: parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}}},
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
		got := parseAndSortResults(animelist.Entry{}, []nyaa.Entry{})
		require.Empty(t, got)
	})
	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}

		got := parseAndSortResults(animelist.Entry{}, input)
		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].Meta.SeasonEpisodeTag, got[i].Meta.SeasonEpisodeTag) <= 0)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := filterEpisodes(animelist.Entry{}, []nyaa.Entry{}, parser.SeasonEpisodeTag{}, animelist.AiringStatusAiring)
		require.Empty(t, got)
	})
	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{}, animelist.AiringStatusAiring)
		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].Meta.SeasonEpisodeTag, got[i].Meta.SeasonEpisodeTag) <= 0)
		}
	})
	t.Run("airing: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}, animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[:1]), got)
	})
	t.Run("airing: with repeated tag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{1}}, animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[0:1]), got)
	})

	t.Run("airing: with latestTag and quality", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03 720p"},
			{Title: "Show3: S03E03 1080p"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}, animelist.AiringStatusAiring)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[1:2]), got)
	})
	t.Run("aired: with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}, animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[:1]), got)
	})
	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{}, animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[2:]), got)
	})
	t.Run("aired: with batch, different qualities", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03 1220x760"},
			{Title: "Show3: S03 1080p"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{}, animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[1:]), got)
	})
	t.Run("aired: with batch, with latestTag", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}, animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[:1]), got)
	})
	t.Run("same tag and quality, different seeders", func(t *testing.T) {
		input := []nyaa.Entry{
			{Title: "Show3: S03E03", Seeders: 1},
			{Title: "Show3: S03E03", Seeders: 10},
			{Title: "Show3: S03"},
		}
		got := filterEpisodes(animelist.Entry{}, input, parser.SeasonEpisodeTag{Season: []int{3}, Episode: []float64{2}}, animelist.AiringStatusAired)
		require.Len(t, got, 1)
		require.Equal(t, parseAndSortResults(animelist.Entry{}, input[1:2]), got)
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
