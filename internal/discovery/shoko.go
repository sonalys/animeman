package discovery

import (
	"context"
	"fmt"
	"path"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/shoko"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/utils/tags"
)

// RunShokoIntegration links shoko's unrecognized files to episodes,
// matching them against the anime list entries and their AniDB ids.
// It drains the shoko queue; only torrents that have fully completed in the
// torrent client are processed, since shoko can only see finished files.
// Torrents still downloading are requeued and retried on the next run.
// It runs on its own one-minute schedule (see runShokoLoop).
func (c *Controller) RunShokoIntegration(
	ctx context.Context,
	entries []animelist.Entry,
) error {
	// Drain the queue first so requeues below don't loop forever.
	pending := make([]torrentsource.Torrent, 0, len(c.shokoQueue))
drain:
	for {
		select {
		case torrentMetadata := <-c.shokoQueue:
			pending = append(pending, torrentMetadata)
		default:
			break drain
		}
	}

	if len(pending) == 0 {
		return nil
	}

	completed, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
		Category:  &c.dep.Config.Category,
		Completed: new(true),
	})
	if err != nil {
		return fmt.Errorf("listing completed torrents: %w", err)
	}
	isCompleted := make(map[string]bool, len(completed))
	for _, torrent := range completed {
		isCompleted[torrent.Hash] = true
	}

	processed := make([]string, 0, len(pending))
	for _, torrentMetadata := range pending {
		if !isCompleted[torrentMetadata.Hash] {
			log.Ctx(ctx).
				Debug().
				Str("torrent", torrentMetadata.Title).
				Msg("skipping torrent: still downloading")
			// Not done yet, retry on the next tick.
			c.enqueueShoko(torrentMetadata)
			continue
		}

		logger := log.Ctx(ctx).With().Str("torrent", torrentMetadata.Title).Logger()
		torrentCtx := logger.WithContext(ctx)

		paths, err := c.dep.TorrentClient.TorrentFiles(torrentCtx, torrentMetadata.Hash)
		if err != nil {
			return fmt.Errorf("listing torrent files: %w", err)
		}
		if len(paths) == 0 {
			logger.Debug().
				Msg("skipping torrent: no files found")
			continue
		}

		for _, filePath := range paths {
			_, err := c.linkShokoFile(
				torrentCtx,
				entries,
				torrentMetadata.Metadata.Title,
				torrentMetadata.Metadata.Tag,
				filePath,
			)
			if err != nil {
				// ponytail: manual linking failures shouldn't kill the whole
				// discovery run, log and move on by dropping the pending tag.
				logger.Warn().
					Err(err).
					Str("file", filePath).
					Msg("failed to link file in shoko, removing pending tag")
			}
		}

		processed = append(processed, torrentMetadata.Hash)
	}

	return nil
}

// linkShokoFile looks a single downloaded file up in shoko. If shoko has not
// run its hash-based AniDB scan yet, we trigger a rescan and leave the file
// for a later run. If the scan already ran and failed to link the file, we
// link it ourselves from the torrent's series/episode tags.
// It reports whether the file is now recognized in shoko.
func (c *Controller) linkShokoFile(
	ctx context.Context,
	entries []animelist.Entry,
	title string,
	tag tags.Tag,
	filePath string,
) (bool, error) {
	// Shoko matches paths by suffix, use the file name only.
	fileName := path.Base(filePath)

	logger := log.Ctx(ctx).With().Str("file", fileName).Logger()
	ctx = logger.WithContext(ctx)

	file, linked, err := c.dep.Shoko.FindFileByPath(ctx, fileName)
	if err != nil {
		return false, fmt.Errorf("finding file by path: %w", err)
	}
	if file == nil {
		logger.Debug().Msg("skipping file: not found in shoko")
		return false, nil
	}
	if linked {
		logger.Debug().Msg("skipping file: already linked in shoko")
		return true, nil
	}

	if !file.Scanned {
		// Shoko hasn't hashed/matched the file yet, ask it to scan and retry
		// on the next run once the scan has completed.
		if err := c.dep.Shoko.RescanFile(ctx, file.ID); err != nil {
			return false, fmt.Errorf("rescanning file: %w", err)
		}
		logger.Debug().Msg("file not scanned by shoko yet, rescan queued")
		return false, nil
	}

	// Shoko scanned the file but couldn't match it by hash. First ask shoko
	// to retry with its local filename-based search, and only fall back to
	// our own linking when that fails too.
	matched, err := c.dep.Shoko.AutoMatchFile(ctx, file.ID)
	if err != nil {
		return false, fmt.Errorf("auto matching file: %w", err)
	}
	if matched {
		logger.Debug().Msg("file matched by shoko local search")
		return true, nil
	}

	// Shoko's local search failed too, link it ourselves from the torrent's
	// series/episode tags.
	linked, err = c.linkFileManually(ctx, entries, file.ID, title, tag)
	if err != nil {
		return false, fmt.Errorf("linking file manually: %w", err)
	}

	return linked, nil
}

// linkFileManually links an unrecognized file to episodes by matching the
// torrent's series tag against the anime list entries and their AniDB ids.
// It reports whether the file was actually linked, so unmatchable files can
// be retried on a later run instead of being marked as recognized.
func (c *Controller) linkFileManually(
	ctx context.Context,
	entries []animelist.Entry,
	fileID int,
	title string,
	tag tags.Tag,
) (bool, error) {
	logger := log.Ctx(ctx)

	entry, ok := matchEntry(title, entries)
	if !ok {
		logger.Debug().
			Str("seriesTag", title).
			Msg("skipping file: no anime list entry matched")
		return false, nil
	}

	anidbID, err := c.findAnidbID(ctx, entry)
	if err != nil {
		return false, fmt.Errorf("finding series: %w", err)
	}
	if anidbID == 0 {
		logger.Debug().Msg("skipping file: series not found in shoko")
		return false, nil
	}

	episodes, err := c.dep.Shoko.FindEpisodes(ctx, anidbID)
	if err != nil {
		return false, fmt.Errorf("finding episodes: %w", err)
	}

	episodeIDs := matchEpisodes(tag, episodes)
	if len(episodeIDs) == 0 {
		logger.Debug().Str("tag", tag.String()).Msg("skipping file: no episode matched")
		return false, nil
	}

	if err := c.dep.Shoko.LinkFileToEpisodes(ctx, fileID, episodeIDs); err != nil {
		return false, fmt.Errorf("linking file: %w", err)
	}

	logger.
		Info().
		Str("title", entry.GetBestTitle()).
		Str("tag", tag.String()).
		Ints("episodeIDs", episodeIDs).
		Msg("linked unrecognized file to episodes")

	return true, nil
}

// findAnidbID resolves the AniDB id of an anime list entry, preferring the
// exact AniList cross-reference known by shoko, and falling back to fuzzy
// title matching when shoko has no such link (or the entry has no AniList id).
func (c *Controller) findAnidbID(ctx context.Context, entry animelist.Entry) (int, error) {
	if entry.AnilistID != 0 {
		anidbID, err := c.dep.Shoko.FindSeriesByAnilistID(ctx, entry.AnilistID)
		if err != nil {
			return 0, fmt.Errorf("finding series by anilist id %d: %w", entry.AnilistID, err)
		}
		if anidbID != 0 {
			return anidbID, nil
		}
	}
	return c.dep.Shoko.FindSeriesByTitle(ctx, entry.GetBestTitle())
}

// matchEntry finds the anime list entry whose titles best match the parsed title.
func matchEntry(title string, entries []animelist.Entry) (animelist.Entry, bool) {
	const minSimilarity = 0.7

	best := -1
	bestScore := 0.0
	for i, entry := range entries {
		for _, entryTitle := range entry.Titles {
			score := utils.CalculateTextSimilarity(entryTitle, title, torrentsource.IgnoreCharset)
			if score > bestScore {
				bestScore = score
				best = i
			}
		}
	}
	if best == -1 || bestScore < minSimilarity {
		return animelist.Entry{}, false
	}
	return entries[best], true
}

// matchEpisodes returns the shoko episode ids matching the tag's episode numbers.
func matchEpisodes(tag tags.Tag, episodes []shoko.Episode) []int {
	byNumber := make(map[int]int, len(episodes))
	for _, episode := range episodes {
		byNumber[episode.Number] = episode.ShokoEpisodeID
	}

	ids := make([]int, 0, len(tag.Episodes))
	for _, episodeNumber := range tag.Episodes {
		// Skip half episodes like 7.5, they have no integer AniDB counterpart.
		if episodeNumber != float64(int(episodeNumber)) {
			continue
		}
		if id, ok := byNumber[int(episodeNumber)]; ok {
			ids = append(ids, id)
		}
	}
	return ids
}
