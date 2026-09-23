package searcher

import (
	"slices"
	"sort"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/mathutils"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/stringutils"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

// prioritize sorts the parsed results by season/episode, title similarity, resolution, release group and seeders.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func prioritize(
	entry animelist.Entry,
	results []torrentsource.Torrent,
	opts torrentsource.SearchOptions,
) []torrentsource.Torrent {
	smallerFunc := func(i, j int) bool {
		first := results[i]
		second := results[j]

		// Sort first by season/episode tag.
		cmp := first.Metadata.Tag.Compare(second.Metadata.Tag)
		if cmp != 0 {
			return cmp < 0
		}

		// Then title similarity.
		titleSimilarityI := mathutils.Max(
			sliceutils.Map(entry.Titles, func(curTitle string) float64 {
				return stringutils.CalculateTextSimilarity(
					curTitle,
					first.Metadata.ShowTitle,
					torrentsource.IgnoreCharset,
				)
			})...)

		titleSimilarityJ := mathutils.Max(
			sliceutils.Map(entry.Titles, func(curTitle string) float64 {
				return stringutils.CalculateTextSimilarity(
					curTitle,
					second.Metadata.ShowTitle,
					torrentsource.IgnoreCharset,
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
