package parser

import (
	"slices"
	"sort"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
)

const IgnoreCharset = " \t!,.:`'\"/\\;-[](){}*【】"

// Prioritize sorts the parsed results by season/episode, title similarity, resolution, release group and seeders.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func Prioritize(
	entry animelist.Entry,
	results []torrentsource.Torrent,
	opts torrentsource.SearchOptions,
) []torrentsource.Torrent {
	parsed := utils.Map(results, func(item torrentsource.Torrent) TorrentMetadata {
		return ParseTorrentMetadata(entry, item, opts.Sources)
	})

	smallerFunc := func(i, j int) bool {
		first := parsed[i]
		second := parsed[j]

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
		return first.Torrent.Seeders > second.Torrent.Seeders
	}

	sort.Slice(results, smallerFunc)

	return results
}
