// Package pager implements the paginated search loop shared by the
// torrent source adapters (nyaa, nekobt): fetch pages until they run out,
// filter/map each page into torrents, and stop early once the source has
// no results newer than the latest downloaded tag.
package pager

import (
	"context"
	"slices"
	"sort"
	"strings"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
)

// Page describes one paginated search over an RSS/torznab feed.
type Page[T any] struct {
	// PageSize is the number of items requested per page.
	PageSize int
	// Fetch fetches one page of raw items at the given offset.
	Fetch func(ctx context.Context, offset int) ([]T, error)
	// Map converts a filtered item into a torrent.
	Map func(T) torrentsource.Torrent
}

// Search runs the paginated loop and returns the mapped torrents.
func (p Page[T]) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	torrents := make([]torrentsource.Torrent, 0, p.PageSize)
	offset := 0

	for {
		items, err := p.Fetch(ctx, offset)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		offset += len(items)
		page := utils.Map(items, p.Map)

		filtered := utils.Filter(page,
			filterSeeders(1),
			filterMetadata(entry, opts.Sources),
			filterSources(opts.Sources),
		)

		if len(items) < p.PageSize || !shouldPaginate(filtered, opts.LatestTag) {
			break
		}
	}

	torrents = prioritize(entry, torrents, opts)

	return torrents, nil
}

// filterSources keeps only torrents whose title contains one of the
// configured sources (release groups), mirroring how the parser extracts
// the release group.
func filterSources(sources []string) func(torrentsource.Torrent) bool {
	return func(item torrentsource.Torrent) bool {
		if len(sources) == 0 {
			return true
		}
		entry := parser.Parse(item.Title, 1, sources)
		return entry.ReleaseGroup != "" && slices.Contains(sources, entry.ReleaseGroup)
	}
}

func filterSeeders(minSeeders int) func(torrentsource.Torrent) bool {
	return func(item torrentsource.Torrent) bool {
		return item.Seeders >= minSeeders
	}
}

func filterMetadata(entry animelist.Entry, sources []string) func(torrentsource.Torrent) bool {
	return func(item torrentsource.Torrent) bool {
		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if !entry.StartDate.IsZero() && item.PublishedAt.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		// If ep number is greater than season ep count, should be removed.
		// This can happen when certain sources mark S2 but use absolute ep number, so they start like S2E13 instead of S2E01.
		// If there's only a single source, then this won't be a problem.
		if len(sources) > 1 && entry.NumEpisodes != 0 {
			metadata := parser.Parse(item.Title, 1, sources)
			if metadata.Tag.FirstEpisode() > float64(entry.NumEpisodes) {
				return false
			}
		}

		return true
	}
}

// shouldPaginate reports whether the source may have more results after this
// page. nekoBT returns newest results first, so paginate while even the
// smallest tag found remains newer than the latest downloaded tag.
func shouldPaginate(items []torrentsource.Torrent, latestTag tags.Tag) bool {
	if latestTag.IsZero() {
		// Nothing downloaded yet: the first page already has everything.
		return false
	}

	if len(items) == 0 {
		return false
	}

	var smallest tags.Tag

	for _, it := range items {
		tag := parser.Parse(it.Title, 1, nil).Tag
		if smallest.IsZero() || tag.Compare(smallest) < 0 {
			smallest = tag
		}
	}

	return smallest.Compare(latestTag) > 0
}

// Prioritize sorts the parsed results by season/episode, title similarity, resolution, release group and seeders.
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
		titleSimilarityI := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				first.Metadata.Title,
				torrentsource.IgnoreCharset,
			)
		})...)

		titleSimilarityJ := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				second.Metadata.Title,
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
