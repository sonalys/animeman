package nyaa

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
)

// Search implements torrentsource.Source.
// It queries the nyaa RSS feed for the entry titles, then filters and sorts
// the results so the newest/best candidates come first.
func (api *API) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	items, err := api.list(ctx, entry, opts)
	if err != nil {
		return nil, fmt.Errorf("listing nyaa: %w", err)
	}

	items = utils.Filter(items,
		filterSeeders(1),
		filterMetadata(entry),
	)

	torrents := utils.Map(items, func(item Item) torrentsource.Torrent {
		pubDate := utils.Must(time.Parse(time.RFC1123Z, item.PubDate))
		return torrentsource.Torrent{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: pubDate,
			Seeders: item.Seeders,
		}
	})

	torrents = parser.Prioritize(entry, torrents, opts)

	return torrents, nil
}

func (api *API) list(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]Item, error) {
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

	return entries, nil
}

func filterSeeders(minSeeders int) func(Item) bool {
	return func(item Item) bool {
		return item.Seeders >= minSeeders
	}
}

// filterMetadata ensures that only coherent and expected nyaa entries are considered for download.
// This function avoids downloading unrelated torrents.
func filterMetadata(
	entry animelist.Entry,
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
				parser.IgnoreCharset,
			) {
				return true
			}
		}

		return false
	}
}
