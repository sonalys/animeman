package discovery

import (
	"context"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/nyaa"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/genericmath"
	"github.com/sonalys/animeman/internal/utils/parser"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	utils "github.com/sonalys/animeman/internal/utils/stringutils"
	"github.com/sonalys/animeman/internal/utils/tags"
	"google.golang.org/grpc/codes"
)

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in Nyaa.si
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	t1 := time.Now()

	log.
		Debug().
		Msgf("discovery started")

	ctx = log.Logger.WithContext(ctx)

	if err := c.TorrentRegenerateTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}

	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}

	for _, entry := range entries {
		if !entry.StartDate.IsZero() && entry.StartDate.After(time.Now()) {
			log.
				Debug().
				Strs("title", entry.Titles).
				Msgf("skipping entry as it hasn't aired yet")
			continue
		}

		log.
			Trace().
			Str("title", selectIdealTitle(entry.Titles)).
			Msgf("starting discovery for entry")

		if err := c.DiscoverEntry(ctx, entry); apperr.Code(err) == codes.Unauthenticated || apperr.Code(err) == codes.Aborted {
			return fmt.Errorf("failed to digest entry: %w", err)
		}

		log.
			Debug().
			Str("title", selectIdealTitle(entry.Titles)).
			Msgf("discovery finished for entry")
	}

	log.
		Debug().
		Int("entries", len(entries)).
		Dur("duration", time.Since(t1)).
		Msg("discovery finished")

	return nil
}

// filterEpisodes will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func filterEpisodes(results []parser.ParsedNyaa, initialTag tags.Tag) []parser.ParsedNyaa {
	out := make([]parser.ParsedNyaa, 0, len(results))

	latestTag := initialTag

	for _, nyaaEntry := range results {
		currentTag := nyaaEntry.ExtractedMetadata.Tag

		if !latestTag.IsZero() {
			if latestTag.IsMultiEpisode() && latestTag.Contains(currentTag) {
				continue
			}

			// This scenario can happen when we are filtering for batches, and the subsequent batch contains the previous batch.
			// Example: S01E01-13, followed by S01.
			// This happens because S01E01-13 < S01, so S01 comes afterwards. But S01 contains the previous tag.
			if currentTag.IsMultiEpisode() && currentTag.Contains(latestTag) {
				out = sliceutils.Filter(out, func(previous parser.ParsedNyaa) bool { return !currentTag.Contains(previous.ExtractedMetadata.Tag) })
			} else if tagCompare(currentTag, latestTag) <= 0 {
				continue
			}
		}

		latestTag = currentTag
		out = append(out, nyaaEntry)
	}

	return out
}

func parseResults(results []nyaa.Item) []parser.ParsedNyaa {
	return sliceutils.Map(results, func(entry nyaa.Item) parser.ParsedNyaa { return parser.NewParsedNyaa(entry) })
}

// sortResults will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func sortResults(entry domain.Entry, results []parser.ParsedNyaa) []parser.ParsedNyaa {
	smallerFunc := func(i, j int) bool {
		first := results[i]
		second := results[j]

		// Sort first by season/episode tag.
		cmp := tagCompare(first.ExtractedMetadata.Tag, second.ExtractedMetadata.Tag)
		if cmp != 0 {
			return cmp < 0
		}

		const ignoreCharset = ",.;:-()[]'`\""

		// Then title similarity.
		titleSimilarityI := genericmath.Max(sliceutils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(curTitle, first.ExtractedMetadata.Title, ignoreCharset)
		})...)

		titleSimilarityJ := genericmath.Max(sliceutils.Map(entry.Titles, func(curTitle string) float64 {
			return utils.CalculateTextSimilarity(curTitle, second.ExtractedMetadata.Title, ignoreCharset)
		})...)

		if titleSimilarityI != titleSimilarityJ {
			return titleSimilarityI > titleSimilarityJ
		}

		// Then resolution.
		cmp = second.ExtractedMetadata.VerticalResolution - first.ExtractedMetadata.VerticalResolution
		if cmp != 0 {
			return cmp < 0
		}

		// Then prioritize number of seeds
		return first.NyaaTorrent.Seeders > second.NyaaTorrent.Seeders
	}

	sort.Slice(results, smallerFunc)

	return results
}

// filterRelevantResults is responsible for filtering and ordering the raw Nyaa feed into valid downloadable torrents.
func filterRelevantResults(
	entry domain.Entry,
	results []parser.ParsedNyaa,
	latestTag tags.Tag,
) []parser.ParsedNyaa {
	results = slices.Clone(results)
	// Requires sorted input, since we use tag progression.
	results = sortResults(entry, results)

	if latestTag.IsZero() && entry.AiringStatus == domain.AiringStatusAired {
		batchResults := sliceutils.Filter(results, func(entry parser.ParsedNyaa) bool { return entry.ExtractedMetadata.Tag.IsMultiEpisode() })
		if len(batchResults) > 0 {
			log.
				Debug().
				Msg("batch detected for aired show, prioritizing batch torrent over individual episodes")
		}
	} else {
		// Remove batches when there are latest tags, avoid episode download duplication.
		results = sliceutils.Filter(results, func(entry parser.ParsedNyaa) bool { return !entry.ExtractedMetadata.Tag.IsMultiEpisode() })
	}

	results = filterEpisodes(results, latestTag)

	if len(results) == 0 {
		log.
			Debug().
			Stringer("latestTag", latestTag).
			Msg("no new episodes detected")

		return nil
	}

	log.
		Debug().
		Int("results", len(results)).
		Msg("newer episodes detected")

	return results
}

func (c *Controller) NyaaSearch(ctx context.Context, entry domain.Entry) ([]nyaa.Item, error) {
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
	sanitizedTitles := sliceutils.Map(entry.Titles, func(title string) string { return titleSanitization.Replace(strings.ToLower(title)) })
	sanitizedTitles = slices.Compact(sanitizedTitles)

	entries, err := c.dep.NYAA.List(ctx, nyaa.ListOptions{
		Titles:              sanitizedTitles,
		VerticalResolutions: c.dep.Config.Qualitites,
		Sources:             c.dep.Config.Sources,
	})
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}

	logger.
		Debug().
		Int("results", len(entries)).
		Msg("searched nyaa for torrent candidates")

	if len(entries) == 0 {
		return nil, nil
	}

	entries = sliceutils.Filter(entries,
		filterMetadata(entry),
	)

	if len(entries) == 0 {
		logger.
			Debug().
			Strs("titles", sanitizedTitles).
			Msg("no results passed the metadata filter")
	}

	return entries, nil
}

// DiscoverEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DiscoverEntry(ctx context.Context, entry domain.Entry) error {
	logger := getLogger(ctx)

	searchResults, err := c.NyaaSearch(ctx, entry)
	if err != nil {
		return fmt.Errorf("searching torrent for anime: %w", err)
	}

	// Remove results without seeders.
	torrentResults := sliceutils.Filter(searchResults, func(e nyaa.Item) bool { return e.Seeders > 0 })

	if len(torrentResults) == 0 {
		logger.
			Trace().
			Msg("no seeded torrent results")

		return nil
	}

	latestTag, err := c.findLatestTag(ctx, entry)
	if err != nil {
		return fmt.Errorf("finding latest anime season episode tag: %w", err)
	}

	if !latestTag.IsZero() {
		logger.
			Debug().
			Str("latestTag", latestTag.String()).
			Msg("detected latest tag")
	}

	parsedTorrents := parseResults(torrentResults)
	parsedTorrents = filterRelevantResults(entry, parsedTorrents, latestTag)

	for _, episodeTorrent := range parsedTorrents {
		if err := c.AddTorrentEntry(ctx, entry, episodeTorrent); err != nil {
			return fmt.Errorf("adding torrent to client: %w", err)
		}

		meta := episodeTorrent.ExtractedMetadata

		logger.
			Info().
			Int("verticalResolution", meta.VerticalResolution).
			Str("title", meta.Title).
			Stringer("latestTag", latestTag).
			Stringer("tag", meta.Tag).
			Msg("new episode added")
	}

	return nil
}
