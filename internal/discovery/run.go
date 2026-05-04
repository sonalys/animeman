package discovery

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
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

	scannedCount := 0
	skippedCount := 0

	for _, entry := range entries {
		// Check if this show should be scanned based on adaptive intervals
		if !c.intervalTracker.ShouldScanNow(entry) {
			skippedCount++
			log.
				Trace().
				Str("title", selectIdealTitle(entry.Titles)).
				Time("nextScanAt", c.intervalTracker.GetNextScanTime(entry)).
				Msgf("skipping entry: not due for scan yet")
			continue
		}

		logger := log.Logger.
			With().
			Str("title", selectIdealTitle(entry.Titles)).
			Logger()

		ctx := logger.WithContext(ctx)

		logger.
			Trace().
			Msgf("starting discovery for entry")

		foundNew, err := c.DiscoverEntry(ctx, entry)
		if errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to digest entry: %w", err)
		}

		// Update the interval tracker with the scan results
		nextScanAt := c.intervalTracker.UpdateState(entry, foundNew)

		scannedCount++

		logger.
			Debug().
			Bool("foundNew", foundNew).
			Time("nextScanAt", nextScanAt).
			Msgf("discovery finished for entry")
	}

	log.
		Debug().
		Int("scanned", scannedCount).
		Int("skipped", skippedCount).
		Dur("duration", time.Since(t1)).
		Msg("discovery finished")

	return nil
}

// filterEpisodes will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func filterEpisodes(
	results []parser.ParsedNyaa,
	initialTag tags.Tag,
	filterData *FilterData,
) ([]parser.ParsedNyaa, tags.Tag) {
	out := make([]parser.ParsedNyaa, 0, len(results))

	var latestDetectedTag tags.Tag

	for _, nyaaEntry := range results {
		currentTag := nyaaEntry.ExtractedMetadata.Tag

		if tagCompare(currentTag, initialTag) <= 0 || tagCompare(currentTag, latestDetectedTag) <= 0 {
			filterData.DiscardedMap[DiscardReasonOlderEpisode]++
			continue
		}

		if !latestDetectedTag.IsZero() {
			if latestDetectedTag.IsMultiEpisode() && latestDetectedTag.Contains(currentTag) {
				filterData.DiscardedMap[DiscardReasonOlderEpisode]++
				continue
			}

			// This scenario can happen when we are filtering for batches, and the subsequent batch contains the previous batch.
			// Example: S01E01-13, followed by S01.
			// This happens because S01E01-13 < S01, so S01 comes afterwards. But S01 contains the previous tag.
			if currentTag.IsMultiEpisode() && currentTag.Contains(latestDetectedTag) {
				out = utils.Filter(out, func(previous parser.ParsedNyaa) bool {
					if currentTag.Contains(previous.ExtractedMetadata.Tag) {
						filterData.DiscardedMap[DiscardReasonOlderEpisode]++
						return false
					}

					return true
				})
			}
		}

		latestDetectedTag = currentTag
		out = append(out, nyaaEntry)
	}

	return out, latestDetectedTag
}

func parseResults(entry animelist.Entry, results []nyaa.Item) []parser.ParsedNyaa {
	return utils.Map(results, func(item nyaa.Item) parser.ParsedNyaa {
		return parser.NewParsedNyaa(entry, item)
	})
}

// sortResults will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func sortResults(entry animelist.Entry, results []parser.ParsedNyaa) []parser.ParsedNyaa {
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
			return utils.CalculateTextSimilarity(curTitle, first.ExtractedMetadata.Title, ignoreCharset)
		})...)

		titleSimilarityJ := utils.Max(utils.Map(entry.Titles, func(curTitle string) float64 {
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
	entry animelist.Entry,
	results []parser.ParsedNyaa,
	latestTag tags.Tag,
	filterData *FilterData,
) []parser.ParsedNyaa {
	results = slices.Clone(results)
	// Requires sorted input, since we use tag progression.
	results = sortResults(entry, results)

	if latestTag.IsZero() && entry.AiringStatus == animelist.AiringStatusAired {
		batchResults := utils.Filter(results, func(entry parser.ParsedNyaa) bool {
			return entry.ExtractedMetadata.Tag.IsMultiEpisode()
		})
		if len(batchResults) > 0 {
			filterData.DiscardedMap[DiscardReasonNotBatch] += uint(len(batchResults))
			results = batchResults
		}
	} else {
		// Remove batches when there are latest tags, avoid episode download duplication.
		results = utils.Filter(results, func(entry parser.ParsedNyaa) bool {
			return !entry.ExtractedMetadata.Tag.IsMultiEpisode()
		})
	}

	results, latestDetectedTag := filterEpisodes(results, latestTag, filterData)
	filterData.LatestFoundTag = latestDetectedTag

	return results
}

type (
	DiscardReason string

	FilterData struct {
		LatestTag      tags.Tag
		LatestFoundTag tags.Tag
		SearchCount    int
		NewCount       int
		DiscardedMap   map[DiscardReason]uint
	}
)

const (
	DiscardReasonNotBatch              DiscardReason = "not_batch"
	DiscardReasonNoSeeder              DiscardReason = "no_seeder"
	DiscardReasonOlderEpisode          DiscardReason = "older_episode"
	DiscardReasonPublishedDateMismatch DiscardReason = "publish_date_mismatch"
	DiscardReasonEpisodeCountMismatch  DiscardReason = "episode_count_mismatch"
	DiscardReasonTitleMismatch         DiscardReason = "title_mismatch"
)

func (c *Controller) NyaaSearch(
	ctx context.Context,
	entry animelist.Entry,
	filterData *FilterData,
) ([]nyaa.Item, error) {
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
	sanitizedTitles = slices.Compact(sanitizedTitles)

	entries, err := c.dep.NYAA.List(ctx, nyaa.ListOptions{
		SearchSuffix:        c.dep.Config.SearchSuffix,
		Titles:              sanitizedTitles,
		VerticalResolutions: c.dep.Config.Qualitites,
		Sources:             c.dep.Config.Sources,
	})
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}

	filterData.SearchCount = len(entries)

	if len(entries) == 0 {
		return nil, nil
	}

	entries = utils.Filter(entries,
		filterMetadata(entry, filterData),
	)

	if len(entries) == 0 {
		logger.
			Debug().
			Msg("no results passed the metadata filter")
	}

	return entries, nil
}

// DiscoverEntry receives an anime list entry and fetches the anime feed, looking for new content.
// It returns the latest discovered tag, whether new episodes were found, and any error.
func (c *Controller) DiscoverEntry(ctx context.Context, entry animelist.Entry) (bool, error) {
	logger := getLogger(ctx)

	filterData := &FilterData{
		SearchCount:  0,
		DiscardedMap: make(map[DiscardReason]uint),
	}

	logger = logger.
		With().
		Any("filterData", filterData).
		Logger()

	searchResults, err := c.NyaaSearch(ctx, entry, filterData)
	if err != nil {
		return false, fmt.Errorf("searching torrent for anime: %w", err)
	}

	// Remove results without seeders.
	torrentResults := utils.Filter(searchResults,
		func(e nyaa.Item) bool {
			if e.Seeders == 0 {
				filterData.DiscardedMap[DiscardReasonNoSeeder]++
				return false
			}

			return true
		},
	)

	if len(torrentResults) == 0 {
		logger.
			Debug().
			Msg("entry discovery stopped: no valid torrent results found")

		return false, nil
	}

	latestTag, err := c.findLatestTag(ctx, entry)
	if err != nil {
		return false, fmt.Errorf("finding latest anime season episode tag: %w", err)
	}

	filterData.LatestTag = latestTag

	parsedTorrents := parseResults(entry, torrentResults)
	parsedTorrents = filterRelevantResults(entry, parsedTorrents, latestTag, filterData)

	foundNewEpisodes := len(parsedTorrents) > 0

	for _, episodeTorrent := range parsedTorrents {
		if err := c.AddTorrentEntry(ctx, entry, episodeTorrent); err != nil {
			return false, fmt.Errorf("adding torrent to client: %w", err)
		}
	}

	filterData.NewCount = len(parsedTorrents)

	logger.
		Info().
		Msg("entry discovery finished")

	return foundNewEpisodes, nil
}
