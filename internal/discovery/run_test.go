package discovery

import (
	"testing"
	"time"

	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/stretchr/testify/require"
)

func Test_buildTaggedNyaaList(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got := parseResults(animelist.Entry{}, []nyaa.Item{})
		got = sortResults(animelist.Entry{}, got)
		require.Empty(t, got)
	})

	t.Run("sort by tag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E01"},
			{Title: "Show3: S03"},
		}

		got := parseResults(animelist.Entry{}, input)
		got = sortResults(animelist.Entry{}, got)

		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.True(
				t,
				tagCompare(got[i-1].ExtractedMetadata.Tag, got[i].ExtractedMetadata.Tag) <= 0,
			)
		}
	})

	t.Run("sort by seeds", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E01", Seeders: 1},
			{Title: "Show3: S03E01", Seeders: 3},
			{Title: "Show3: S03E01", Seeders: 2},
		}

		got := parseResults(animelist.Entry{}, input)
		got = sortResults(animelist.Entry{}, got)

		require.Len(t, got, len(input))

		for i := 1; i < len(got); i++ {
			require.LessOrEqual(t, got[i].NyaaTorrent.Seeders, got[i-1].NyaaTorrent.Seeders)
		}
	})
}

func Test_filterNyaaFeed(t *testing.T) {
	newEntry := func(airingStatus animelist.AiringStatus) animelist.Entry {
		return animelist.NewEntry(
			nil,
			animelist.ListStatusWatching,
			airingStatus,
			time.Now(),
			time.Now(),
			0,
			nil,
		)
	}

	t.Run("empty", func(t *testing.T) {
		got := filterRelevantResults(
			animelist.Entry{},
			[]parser.ParsedNyaa{},
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)
		require.Empty(t, got)
	})

	t.Run("airing: no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03E01"},
		}

		parsed := parseResults(animelist.Entry{}, input)
		got := filterRelevantResults(
			animelist.Entry{},
			parsed,
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)

		require.Len(t, got, len(input))
		for i := 1; i < len(got); i++ {
			require.True(
				t,
				tagCompare(got[i-1].ExtractedMetadata.Tag, got[i].ExtractedMetadata.Tag) <= 0,
			)
		}
	})

	t.Run("aired: with batch, no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E03"},
			{Title: "Show3: S03E02"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(animelist.Entry{}, input)
		got := filterRelevantResults(
			newEntry(animelist.AiringStatusAired),
			parsed,
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)

		require.Equal(t, parsed[2:], got)
	})

	t.Run("aired: with batch and multi episode, no latestTag", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03E01-13"},
			{Title: "Show3: S03"},
		}

		parsed := parseResults(animelist.Entry{}, input)
		got := filterRelevantResults(
			newEntry(animelist.AiringStatusAired),
			parsed,
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)

		require.Equal(t, parsed[1:], got)
	})

	t.Run("aired: with batch, different qualities", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S03 1220x760"},
			{Title: "Show3: S03 1080p"},
		}

		parsed := parseResults(animelist.Entry{}, input)
		got := filterRelevantResults(
			newEntry(animelist.AiringStatusAired),
			parsed,
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)

		require.Len(t, got, 1)
		require.Equal(t, parsed[1:], got)
	})

	t.Run("batch for different seasons", func(t *testing.T) {
		input := []nyaa.Item{
			{Title: "Show3: S2"},
			{Title: "Show3: S1"},
			{Title: "Show3: S3"},
		}

		parsed := parseResults(animelist.Entry{}, input)
		got := filterRelevantResults(
			newEntry(animelist.AiringStatusAired),
			parsed,
			tags.Zero,
			&FilterData{DiscardReason: make(map[DiscardReason]uint)},
		)

		require.Len(t, got, 3)
	})
}
