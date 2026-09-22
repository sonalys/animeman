package torrentsource

import (
	"context"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
)

// Torrent is a normalized torrent entry returned by any torrent source.
type Torrent struct {
	Title       string
	Link        string
	Seeders     int
	Hash        string
	PublishedAt time.Time

	Metadata parser.Metadata
}

// SearchOptions controls how a source searches for torrents of an entry.
type SearchOptions struct {
	SearchSuffix string
	Sources      []string
	Qualities    []string
	// LatestTag is the newest season/episode tag already downloaded in the
	// torrent client. Sources may paginate past it when the first page only
	// contains older episodes.
	LatestTag tags.Tag
}

// Source is the port implemented by every torrent source (nyaa, nekobt, ...).
// It receives an anime list entry and returns downloadable torrents for it,
// already filtered and sorted so the newest/best candidates come first.
type Source interface {
	Search(ctx context.Context, entry animelist.Entry, opts SearchOptions) ([]Torrent, error)
}

// Prioritize sorts the parsed results by season/episode, title similarity, resolution, release group and seeders.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func Prioritize(
	entry animelist.Entry,
	results []Torrent,
	opts SearchOptions,
) []Torrent {
	smallerFunc := func(i, j int) bool {
		first := results[i]
		second := results[j]

		// Sort first by season/episode tag.
		cmp := first.Metadata.Tag.Compare(second.Metadata.Tag)
		if cmp != 0 {
			return cmp < 0
		}

		// Then title similarity.
		titleSimilarityI := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				first.Metadata.Title,
				IgnoreCharset,
			)
		})...)

		titleSimilarityJ := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				second.Metadata.Title,
				IgnoreCharset,
			)
		})...)

		if titleSimilarityI != titleSimilarityJ {
			return titleSimilarityI > titleSimilarityJ
		}

		// Then quality, e.g. "1080 HEVC" beats "1080" or "720".
		// Lower index = higher priority, so a higher quality always wins.
		if len(opts.Qualities) > 0 {
			cmp = qualityScore(first.Title, opts.Qualities) -
				qualityScore(second.Title, opts.Qualities)
			if cmp != 0 {
				return cmp < 0
			}
		}

		// Then resolution.
		cmp = second.Metadata.VerticalResolution - first.Metadata.VerticalResolution
		if cmp != 0 {
			return cmp < 0
		}

		// Then source.
		if len(opts.Sources) > 0 &&
			first.Metadata.ReleaseGroup != second.Metadata.ReleaseGroup {
			cmp = slices.Index(
				opts.Sources,
				first.Metadata.ReleaseGroup,
			) - slices.Index(
				opts.Sources,
				second.Metadata.ReleaseGroup,
			)
			if cmp != 0 {
				return cmp < 0
			}
		}

		// Then prioritize number of seeds
		return first.Seeders > second.Seeders
	}

	sort.Slice(results, smallerFunc)

	return results
}

const IgnoreCharset = " \t!,.:`'\"/\\;-[](){}*【】"

// qualityScore returns the index of the first quality that matches the title, or len(qualities) if none matched.
func qualityScore(title string, qualities []string) int {
	normalizedTitle := strings.ToLower(title)

	for index, quality := range qualities {
		matched := 0
		keywords := strings.Fields(quality)
		for keyword := range strings.FieldsSeq(quality) {
			if strings.Contains(normalizedTitle, strings.ToLower(keyword)) {
				matched++
			}
		}
		// Only count qualities where every keyword matched.
		if matched == len(keywords) {
			return index
		}
	}
	return len(qualities)
}
