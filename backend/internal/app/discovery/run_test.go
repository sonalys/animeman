package discovery

import (
	"reflect"
	"testing"
	"time"

	"github.com/sonalys/animeman/internal/adapters/nyaa"
	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/parser"
	"github.com/sonalys/animeman/internal/utils/tags"
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
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(1, 16)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(1, 17)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(1, 18)}},
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
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{
				{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
				{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
			},
		},
		{
			name: "tag",
			args: args{
				latestTag: tags.SeasonEpisode(3, 1),
				list: []parser.ParsedNyaa{
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{
				{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
			},
		},
		{
			name: "season batch",
			args: args{
				latestTag: tags.Tag{Seasons: []int{3}},
				list: []parser.ParsedNyaa{
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
		{
			name: "same tag",
			args: args{
				latestTag: tags.SeasonEpisode(3, 2),
				list: []parser.ParsedNyaa{
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 1)}},
					{ExtractedMetadata: parser.Metadata{Tag: tags.SeasonEpisode(3, 2)}},
				},
			},
			want: []parser.ParsedNyaa{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := filterEpisodes(tt.args.list, tt.args.latestTag); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("filterEpisodes() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseResults([]nyaa.Item{})
		got = sortResults(domain.Entry{}, got)
		require.Empty(t, got)
	})

	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}

		got := parseResults(input)
		got = sortResults(domain.Entry{}, got)

		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].ExtractedMetadata.Tag, got[i].ExtractedMetadata.Tag) <= 0)
		}
	})

	t.Run("sort by seeds", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E01", Seeders: 1},
			{Title: "Show3: S03E01", Seeders: 3},
			{Title: "Show3: S03E01", Seeders: 2},
		}

		got := parseResults(input)
		got = sortResults(domain.Entry{}, got)

		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.LessOrEqual(t, got[i].NyaaTorrent.Seeders, got[i-1].NyaaTorrent.Seeders)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	newEntry := func(airingStatus domain.AiringStatus) domain.Entry {
		return domain.NewEntry(nil, domain.ListStatusWatching, airingStatus, time.Now(), 0)
	}

	t.Run("empty", func(t *testing.T) {
		got := filterRelevantResults(domain.Entry{}, []parser.ParsedNyaa{}, tags.Zero)
		require.Empty(t, got)
	})

	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(domain.Entry{}, parsed, tags.Zero)

		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(t, tagCompare(got[i-1].ExtractedMetadata.Tag, got[i].ExtractedMetadata.Tag) <= 0)
		}
	})

	t.Run("airing: with latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		entry := newEntry(domain.AiringStatusAiring)
		parsedTorrents := parseResults(input)
		latestTag := tags.SeasonEpisode(3, 2)

		got := filterRelevantResults(entry, parsedTorrents, latestTag)

		require.Equal(t, parsedTorrents[:1], got)
	})

	t.Run("airing: with repeated tag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(domain.Entry{}, parsed, tags.SeasonEpisode(3, 1))

		require.Len(t, got, 1)
		require.Equal(t, parsed[0:1], got)
	})

	t.Run("airing: with latestTag and quality", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03 720p"},
			{Title: "Show3: S03E03 1080p"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(domain.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:2], got)
	})

	t.Run("aired: with latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(domain.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[:1], got)
	})

	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(domain.AiringStatusAired), parsed, tags.Zero)

		require.Equal(t, parsed[2:], got)
	})

	t.Run("aired: with batch and multi episode, no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E01-13"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(domain.AiringStatusAired), parsed, tags.Zero)

		require.Equal(t, parsed[1:], got)
	})

	t.Run("aired: with batch, different qualities", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03 1220x760"},
			{Title: "Show3: S03 1080p"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(domain.AiringStatusAired), parsed, tags.Zero)

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:], got)
	})

	t.Run("aired: with batch, with latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(domain.AiringStatusAired), parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[:1], got)
	})

	t.Run("same tag and quality, different seeders", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03", Seeders: 1},
			{Title: "Show3: S03E03", Seeders: 10},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(domain.Entry{}, parsed, tags.SeasonEpisode(3, 2))

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:2], got)
	})

	t.Run("batch for different seasons", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S2"},
			{Title: "Show3: S1"},
			{Title: "Show3: S3"},
		}

		parsed := parseResults(input)
		got := filterRelevantResults(newEntry(domain.AiringStatusAired), parsed, tags.Zero)

		require.Len(t, got, 3)
	})
}
