package nyaa

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/ports/animelist"
)

const ignoreCharset = " \t!,.:`'\"/\\;-[](){}*【】"

func getLogger(ctx context.Context) zerolog.Logger {
	return *zerolog.Ctx(ctx)
}

// Search implements torrentsource.Source.
// It queries the nyaa RSS feed for the entry titles, then filters and sorts
// the results so the newest/best candidates come first.
func (api *API) Search(ctx context.Context, entry animelist.Entry, opts torrentsource.SearchOptions) ([]torrentsource.Torrent, error) {
	items, err := api.list(ctx, entry, opts)
	if err != nil {
		return nil, fmt.Errorf("listing nyaa: %w", err)
	}

	items = utils.Filter(items, filterMetadata(entry, items))

	parsed := parseResults(entry, items, opts)
	parsed = sortResults(entry, parsed, opts)

	return toTorrents(parsed), nil
}

func (api *API) list(ctx context.Context, entry animelist.Entry, opts torrentsource.SearchOptions) ([]Item, error) {
	logger := getLogger(ctx)

	titleSanitization := strings.NewReplacer(
		"-", " ",
		"\"", " ",
		"'", " ",
		"(", " ",
		")", " ",
	)

	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	sanitizedTitles := utils.Transform(entry.Titles,
		strings.ToLower,
		parser.StripTitle,
		parser.StripSubtitle,
		titleSanitization.Replace,
	)

	sort.Strings(sanitizedTitles)
	sanitizedTitles = slices.Compact(sanitizedTitles)

	entries, err := api.List(ctx, ListOptions{
		SearchSuffix:        opts.SearchSuffix,
		Titles:              sanitizedTitles,
		VerticalResolutions: opts.Qualities,
		Sources:             opts.Sources,
	})
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}

	if len(entries) == 0 {
		return nil, nil
	}

	entries = utils.Filter(entries, filterMetadata(entry, nil))

	if len(entries) == 0 {
		logger.Debug().Msg("no results passed the metadata filter")
	}

	return entries, nil
}

// filterMetadata ensures that only coherent and expected nyaa entries are considered for download.
// This function avoids downloading unrelated torrents.
func filterMetadata(
	entry animelist.Entry,
	_ []Item,
) func(e Item) bool {
	return func(nyaaEntry Item) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, nyaaEntry.PubDate))

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if publishedDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		// Check if nyaa entry episode is greater than the animelist episode count.
		if entry.NumEpisodes > 0 &&
			parser.Parse(nyaaEntry.Title, 1, nil).Tag.LastEpisode() > float64(entry.NumEpisodes) {
			return false
		}

		nyaaTitleWithoutTags := parser.StripTags(nyaaEntry.Title)

		for _, originalTitle := range entry.Titles {
			// Remove season information from the original title, as it is not always present in the nyaa entry.
			originalTitleWithoutSeason := parser.StripSeason(originalTitle)
			originalTitleWithoutSubtitle := parser.StripSubtitle(originalTitleWithoutSeason)

			if utils.MatchPrefixFlexible(
				nyaaTitleWithoutTags,
				originalTitleWithoutSubtitle,
				ignoreCharset,
			) {
				return true
			}
		}

		return false
	}
}

func parseResults(entry animelist.Entry, results []Item, opts torrentsource.SearchOptions) []parser.ParsedNyaa {
	return utils.Map(results, func(item Item) parser.ParsedNyaa {
		return parser.NewParsedNyaa(entry, toTorrent(item), opts.Sources)
	})
}

// sortResults sorts the parsed results by season/episode, title similarity, resolution, release group and seeders.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func sortResults(
	entry animelist.Entry,
	results []parser.ParsedNyaa,
	opts torrentsource.SearchOptions,
) []parser.ParsedNyaa {
	smallerFunc := func(i, j int) bool {
		first := results[i]
		second := results[j]

		// Sort first by season/episode tag.
		cmp := tagCompare(first.ExtractedMetadata.Tag, second.ExtractedMetadata.Tag)
		if cmp != 0 {
			return cmp < 0
		}

		// Then title similarity.
		titleSimilarityI := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				first.ExtractedMetadata.Title,
				ignoreCharset,
			)
		})...)

		titleSimilarityJ := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(
				curTitle,
				second.ExtractedMetadata.Title,
				ignoreCharset,
			)
		})...)

		if titleSimilarityI != titleSimilarityJ {
			return titleSimilarityI > titleSimilarityJ
		}

		// Then resolution.
		cmp = second.ExtractedMetadata.VerticalResolution - first.ExtractedMetadata.VerticalResolution
		if cmp != 0 {
			return cmp < 0
		}

		// Then source.
		if len(opts.Sources) > 0 &&
			first.ExtractedMetadata.ReleaseGroup != second.ExtractedMetadata.ReleaseGroup {
			cmp = slices.Index(
				opts.Sources,
				first.ExtractedMetadata.ReleaseGroup,
			) - slices.Index(
				opts.Sources,
				second.ExtractedMetadata.ReleaseGroup,
			)
			if cmp != 0 {
				return cmp < 0
			}
		}

		// Then prioritize number of seeds
		return first.NyaaTorrent.Seeders > second.NyaaTorrent.Seeders
	}

	sort.Slice(results, smallerFunc)

	return results
}

func toTorrent(item Item) torrentsource.Torrent {
	pubDate := utils.Must(time.Parse(time.RFC1123Z, item.PubDate))
	return torrentsource.Torrent{
		Title:   item.Title,
		Link:    item.Link,
		PubDate: pubDate,
		Seeders: item.Seeders,
	}
}

func toTorrents(parsed []parser.ParsedNyaa) []torrentsource.Torrent {
	return utils.Map(parsed, func(p parser.ParsedNyaa) torrentsource.Torrent {
		return p.NyaaTorrent
	})
}

// tagCompare receives 2 series tags, Example: S02E01 and S02E02.
// it will return the comparison of Tag1, Tag2.
// -1 = Tag1 < Tag2.
// 0 = Tag1 == Tag2.
// 1 = Tag1 > Tag2.
func tagCompare(a, b tags.Tag) int {
	if a.LastSeason() < b.LastSeason() {
		return -1
	}

	if a.LastSeason() > b.LastSeason() {
		return 1
	}

	if a.FirstSeason() < b.FirstSeason() {
		return 1
	}

	if a.FirstSeason() > b.FirstSeason() {
		return -1
	}

	if a.IsMultiEpisode() && !b.IsMultiEpisode() {
		return 1
	}

	if b.IsMultiEpisode() && !a.IsMultiEpisode() {
		return -1
	}

	aLastEpisode := a.LastEpisode()
	bLastEpisode := b.LastEpisode()

	if aLastEpisode == bLastEpisode {
		return 0
	}

	if aLastEpisode == 0 {
		return 1
	}

	if bLastEpisode == 0 {
		return -1
	}

	if a.LastEpisode() < b.LastEpisode() {
		return -1
	}

	if a.LastEpisode() > b.LastEpisode() {
		return 1
	}

	return 0
}
