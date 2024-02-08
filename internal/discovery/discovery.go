package discovery

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// ParsedNyaa holds a parsed entry from Nyaa.
// Used for smart episode detection.
type ParsedNyaa struct {
	// Metadata parsed from title.
	meta parser.Metadata
	// Example: S02E03.
	seasonEpisodeTag string
	// Nyaa entry.
	entry nyaa.Entry
}

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in Nyaa.si
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	t1 := time.Now()
	log.Debug().Msgf("discovery started")
	entries, err := c.dep.AnimeListClient.GetCurrentlyWatching(ctx)
	if err != nil {
		return fmt.Errorf("fetching anime list: %w", err)
	}
	var totalCount int
	for _, entry := range entries {
		count, err := c.DigestAnimeListEntry(ctx, entry)
		if err != nil {
			if errors.Is(err, torrentclient.ErrUnauthorized) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("failed to digest entry: %w", err)
			}
			continue
		}
		totalCount += count
	}
	log.Info().Int("animeListCount", len(entries)).Int("addedCount", totalCount).Str("duration", time.Since(t1).String()).Msgf("discovery finished")
	return nil
}

// findNyaaBatch filters Nyaa entries for a single Batch entry.
func findNyaaBatch(entries []nyaa.Entry) []nyaa.Entry {
	for _, entry := range entries {
		if meta := parser.TitleParse(entry.Title); meta.IsMultiEpisode {
			return []nyaa.Entry{entry}
		}
	}
	return entries
}

// parseNyaaEntries will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func parseNyaaEntries(entries []nyaa.Entry) []ParsedNyaa {
	out := make([]ParsedNyaa, 0, len(entries))
	for _, entry := range entries {
		meta := parser.TitleParse(entry.Title)
		out = append(out, ParsedNyaa{
			meta:             meta,
			seasonEpisodeTag: meta.TagBuildSeasonEpisode(),
			entry:            entry,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return tagCompare(out[i].seasonEpisodeTag, out[j].seasonEpisodeTag) < 0
	})
	return out
}

// episodeFilter will only return ParsedNyaa entries that are more recent than the given latestTag.
// excludeBatch is used when a show is airing or you have already downloaded some episodes of the season.
// excludeBatch avoids downloading a batch for episodes which you already have.
func episodeFilter(list []ParsedNyaa, latestTag string, excludeBatch bool) []ParsedNyaa {
	out := make([]ParsedNyaa, 0, len(list))
	for _, nyaaEntry := range list {
		if excludeBatch && nyaaEntry.meta.IsMultiEpisode {
			continue
		}
		// Make sure we only add episodes ahead of the current ones in the qBittorrent.
		if tagCompare(nyaaEntry.seasonEpisodeTag, latestTag) <= 0 {
			continue
		}
		latestTag = nyaaEntry.seasonEpisodeTag
		out = append(out, nyaaEntry)
	}
	return out
}

// filterNyaaFeed is responsible for filtering and ordering the raw Nyaa feed into valid downloadable torrents.
func filterNyaaFeed(entries []nyaa.Entry, latestTag string, animeStatus animelist.AiringStatus) []ParsedNyaa {
	// If we don't have any episodes, and show is released, try to find a batch for all episodes.
	useBatch := latestTag == "" && animeStatus == animelist.AiringStatusAired
	if useBatch {
		entries = findNyaaBatch(entries)
	}
	return episodeFilter(parseNyaaEntries(entries), latestTag, !useBatch)
}

func ForEach[T any](in []T, f func(T) T) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		out = append(out, f(in[i]))
	}
	return out
}

// DigestAnimeListEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DigestAnimeListEntry(ctx context.Context, entry animelist.Entry) (count int, err error) {
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	titleQuery := nyaa.OrQuery(ForEach(entry.Titles, func(title string) string { return parser.TitleStrip(title) }))
	sourceQuery := nyaa.OrQuery(c.dep.Config.Sources)
	qualityQuery := nyaa.OrQuery(c.dep.Config.Qualitites)

	nyaaEntries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	log.Debug().Str("entry", entry.Titles[0]).Msgf("found %d torrents", len(nyaaEntries))
	if err != nil {
		return count, fmt.Errorf("getting nyaa list: %w", err)
	}
	// There should always be torrents for entries, if there aren't we can just exit the routine.
	if len(nyaaEntries) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.Titles[0])
		return count, nil
	}
	latestTag, err := c.TagGetLatest(ctx, entry)
	if err != nil {
		return count, fmt.Errorf("getting latest tag: %w", err)
	}
	for _, nyaaEntry := range filterNyaaFeed(nyaaEntries, latestTag, entry.AiringStatus) {
		if err := c.DigestNyaaTorrent(ctx, entry, nyaaEntry); err != nil {
			log.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
		count++
	}
	return count, nil
}
