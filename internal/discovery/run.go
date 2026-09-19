package discovery

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
)

const ignoreCharset = " \t!,.:`'\"/\\;-[](){}*【】"

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in the torrent source.
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	t1 := time.Now()

	log.
		Debug().
		Msgf("discovery started")

	ctx = log.Logger.WithContext(ctx)

	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}

	if err := c.TorrentRegenerateTags(ctx, entries); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
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

// filterEpisodes will only return entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func filterEpisodes(
	results []parser.TorrentMetadata,
	initialTag tags.Tag,
) ([]parser.TorrentMetadata, tags.Tag) {
	out := make([]parser.TorrentMetadata, 0, len(results))

	var latestDetectedTag tags.Tag

	for _, nyaaEntry := range results {
		currentTag := nyaaEntry.Metadata.Tag

		if tagCompare(currentTag, initialTag) <= 0 ||
			tagCompare(currentTag, latestDetectedTag) <= 0 {
			continue
		}

		if !latestDetectedTag.IsZero() {
			if latestDetectedTag.IsMultiEpisode() && latestDetectedTag.Contains(currentTag) {
				continue
			}

			// This scenario can happen when we are filtering for batches, and the subsequent batch contains the previous batch.
			// Example: S01E01-13, followed by S01.
			// This happens because S01E01-13 < S01, so S01 comes afterwards. But S01 contains the previous tag.
			if currentTag.IsMultiEpisode() && currentTag.Contains(latestDetectedTag) {
				out = utils.Filter(out, func(previous parser.TorrentMetadata) bool {
					return !currentTag.Contains(previous.Metadata.Tag)
				})
			}
		}

		latestDetectedTag = currentTag
		out = append(out, nyaaEntry)
	}

	return out, latestDetectedTag
}

// filterRelevantResults is responsible for filtering and ordering the raw feed into valid downloadable torrents.
func filterRelevantResults(
	entry animelist.Entry,
	results []parser.TorrentMetadata,
	latestTag tags.Tag,
) []parser.TorrentMetadata {
	results = slices.Clone(results)

	if latestTag.IsZero() && entry.AiringStatus == animelist.AiringStatusAired {
		batchResults := utils.Filter(results, func(entry parser.TorrentMetadata) bool {
			return entry.Metadata.Tag.IsMultiEpisode()
		})
		if len(batchResults) > 0 {
			results = batchResults
		}
	} else {
		// Remove batches when there are latest tags, avoid episode download duplication.
		results = utils.Filter(results, func(entry parser.TorrentMetadata) bool {
			return !entry.Metadata.Tag.IsMultiEpisode()
		})
	}

	results, _ = filterEpisodes(results, latestTag)

	return results
}

// DiscoverEntry receives an anime list entry and fetches the anime feed, looking for new content.
// It returns the latest discovered tag, whether new episodes were found, and any error.
func (c *Controller) DiscoverEntry(ctx context.Context, entry animelist.Entry) (bool, error) {
	logger := getLogger(ctx)

	results, err := c.dep.Source.Search(ctx, entry, torrentsource.SearchOptions{
		SearchSuffix: c.dep.Config.SearchSuffix,
		Sources:      c.dep.Config.ReleaseGroups,
		Qualities:    c.dep.Config.Qualitites,
	})
	if err != nil {
		return false, fmt.Errorf("searching torrent for anime: %w", err)
	}

	if len(results) == 0 {
		logger.
			Debug().
			Msg("entry discovery stopped: no valid torrent results found")

		return false, nil
	}

	latestTag, err := c.findLatestTag(ctx, entry)
	if err != nil {
		return false, fmt.Errorf("finding latest anime season episode tag: %w", err)
	}

	parsedTorrents := parseResults(entry, results, c.dep.Config)
	parsedTorrents = filterRelevantResults(
		entry,
		parsedTorrents,
		latestTag,
	)

	foundNewEpisodes := len(parsedTorrents) > 0

	for _, episodeTorrent := range parsedTorrents {
		if err := c.AddTorrentEntry(ctx, entry, episodeTorrent); err != nil {
			return false, fmt.Errorf("adding torrent to client: %w", err)
		}
	}

	logger.
		Info().
		Int("newCount", len(parsedTorrents)).
		Any("latestTag", latestTag).
		Msg("entry discovery finished")

	return foundNewEpisodes, nil
}

func parseResults(
	entry animelist.Entry,
	results []torrentsource.Torrent,
	config Config,
) []parser.TorrentMetadata {
	return utils.Map(results, func(item torrentsource.Torrent) parser.TorrentMetadata {
		return parser.ParseTorrentMetadata(entry, item, config.ReleaseGroups)
	})
}
