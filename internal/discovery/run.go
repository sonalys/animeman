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

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in the torrent source.
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	ctx, cancel := context.WithTimeout(ctx, 5*time.Minute)
	defer cancel()

	t1 := time.Now()

	log.
		Debug().
		Msgf("discovery started")

	ctx = log.Logger.WithContext(ctx)

	entries, err := c.dep.AnimeListSource.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}

	// MAL-only entries get their AniList id backfilled, so shoko can be
	// matched by exact id instead of fuzzy title search.
	if c.dep.Shoko != nil {
		for i := range entries {
			entry := &entries[i]

			if entry.AnilistID != 0 || entry.MALID == 0 {
				continue
			}

			anilistID, err := c.dep.AnilistIDResolver.ResolveMAL(ctx, entry.MALID)
			if err != nil {
				log.Ctx(ctx).
					Warn().
					Err(err).
					Int("malID", entry.MALID).
					Msg("failed to resolve anilist id")
				continue
			}

			entry.
				WithAnilistID(anilistID).
				WithMALID(entry.MALID)
		}
	}

	if err := c.TorrentRegenerateTags(ctx, entries); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}

	var addedTorrents []parser.TorrentMetadata

	scannedCount := 0
	skippedCount := 0

	for _, entry := range entries {
		// Check if this show should be scanned based on adaptive intervals
		if !c.intervalTracker.shouldScanNow(entry) {
			skippedCount++
			log.
				Trace().
				Str("title", entry.GetBestTitle()).
				Time("nextScanAt", c.intervalTracker.getNextScanTime(entry)).
				Msgf("skipping entry: not due for scan yet")
			continue
		}

		logger := log.Logger.
			With().
			Str("title", entry.GetBestTitle()).
			Logger()

		ctx := logger.WithContext(ctx)

		logger.
			Trace().
			Msgf("starting discovery for entry")

		foundNew, added, err := c.DiscoverEntry(ctx, entry)
		if errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
			return fmt.Errorf("failed to digest entry: %w", err)
		}
		addedTorrents = append(addedTorrents, added...)

		// Update the interval tracker with the scan results
		nextScanAt := c.intervalTracker.updateState(entry, foundNew)

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

	if scannedCount == 0 {
		return nil
	}

	if c.dep.Shoko != nil && len(addedTorrents) > 0 {
		if err := c.RunShokoIntegration(ctx, entries, addedTorrents); err != nil {
			return fmt.Errorf("shoko integration: %w", err)
		}
	}

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

		if currentTag.Compare(initialTag) <= 0 ||
			currentTag.Compare(latestDetectedTag) <= 0 {
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
// It returns the latest discovered tag, whether new episodes were found, the torrents added
// to the client (used by the shoko integration), and any error.
func (c *Controller) DiscoverEntry(
	ctx context.Context,
	entry animelist.Entry,
) (bool, []parser.TorrentMetadata, error) {
	logger := getLogger(ctx)

	results, err := c.dep.TorrentSource.Search(ctx, entry, torrentsource.SearchOptions{
		SearchSuffix: c.dep.Config.SearchSuffix,
		Sources:      c.dep.Config.ReleaseGroups,
		Qualities:    c.dep.Config.Qualitites,
	})
	if err != nil {
		return false, nil, fmt.Errorf("searching torrent for anime: %w", err)
	}

	if len(results) == 0 {
		logger.
			Debug().
			Msg("entry discovery stopped: no valid torrent results found")

		return false, nil, nil
	}

	latestTag, err := c.getLatestDownloadedTag(ctx, entry)
	if err != nil {
		return false, nil, fmt.Errorf("finding latest anime season episode tag: %w", err)
	}

	newTorrents := parseResults(entry, results, c.dep.Config)

	for _, parsed := range newTorrents {
		logger.
			Debug().
			Str("torrentTitle", parsed.Torrent.Title).
			Str("parsedTitle", parsed.Metadata.Title).
			Str("tag", parsed.Metadata.Tag.String()).
			Str("seriesTag", parsed.Metadata.BuildSeriesTag()).
			Str("releaseGroup", parsed.Metadata.ReleaseGroup).
			Int("resolution", parsed.Metadata.VerticalResolution).
			Msg("parsed torrent result")
	}

	newTorrents = filterRelevantResults(
		entry,
		newTorrents,
		latestTag,
	)

	for _, parsed := range newTorrents {
		logger.
			Debug().
			Str("torrentTitle", parsed.Torrent.Title).
			Str("tag", parsed.Metadata.Tag.String()).
			Msg("torrent result kept after filtering")
	}

	foundNewEpisodes := len(newTorrents) > 0

	// Hand the added torrents to the shoko integration directly, using the
	// info hashes provided by the torrent source, no listing needed.
	var added []parser.TorrentMetadata
	for i := range newTorrents {
		torrentMetadata := &newTorrents[i]

		hash, err := c.AddTorrentEntry(ctx, entry, *torrentMetadata)
		if err != nil {
			return false, nil, fmt.Errorf("adding torrent to client: %w", err)
		}

		torrentMetadata.Torrent.Hash = hash

		logger.
			Info().
			Str("torrentTitle", torrentMetadata.Torrent.Title).
			Str("tag", torrentMetadata.Metadata.Tag.String()).
			Msg("added torrent to client")

		if c.dep.Shoko != nil && hash != "" {
			added = append(added, *torrentMetadata)
		}
	}

	logger.
		Debug().
		Int("newCount", len(newTorrents)).
		Stringer("latestTag", latestTag).
		Msg("finished entry discovery")

	return foundNewEpisodes, added, nil
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
