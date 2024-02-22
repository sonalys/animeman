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
	"github.com/sonalys/animeman/internal/utils"
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

func NewParsedNyaa(entry nyaa.Entry) ParsedNyaa {
	meta := parser.TitleParse(entry.Title)
	return ParsedNyaa{
		meta:             meta,
		seasonEpisodeTag: meta.TagBuildSeasonEpisode(),
		entry:            entry,
	}
}

// RunDiscovery controls the discovery routine,
// fetching entries from your anime list and looking for updates in Nyaa.si
// After finding updates, it will verify episode collision and dispatch it to your torrent client.
func (c *Controller) RunDiscovery(ctx context.Context) error {
	t1 := time.Now()
	if err := c.UpdateExistingTorrentsTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}
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

// parseNyaaEntries will digest the raw data from Nyaa into a parsed metadata struct `ParsedNyaa`.
// it will also sort the response by season and episode.
// it's important it returns a crescent season/episode list, so you don't download a recent episode and
// don't download the oldest ones in case you don't have all episodes since your latestTag.
func parseNyaaEntries(entries []nyaa.Entry) []ParsedNyaa {
	resp := utils.ForEach(entries, func(entry nyaa.Entry) ParsedNyaa { return NewParsedNyaa(entry) })
	sort.Slice(resp, func(i, j int) bool {
		cmp := tagCompare(resp[i].seasonEpisodeTag, resp[j].seasonEpisodeTag)
		// For same tag, we compare vertical resolution, prioritizing better quality.
		if cmp == 0 {
			return resp[i].meta.VerticalResolution > resp[j].meta.VerticalResolution
		}
		return cmp < 0
	})
	return resp
}

// filterNyaaFeed is responsible for filtering and ordering the raw Nyaa feed into valid downloadable torrents.
func filterNyaaFeed(entries []nyaa.Entry, latestTag string, animeStatus animelist.AiringStatus) []ParsedNyaa {
	// If we don't have any episodes, and show is released, try to find a batch for all episodes.
	useBatch := latestTag == "" && animeStatus == animelist.AiringStatusAired
	if useBatch {
		batchEntry, ok := utils.Find(entries, func(e nyaa.Entry) bool { return parser.TitleParse(e.Title).IsMultiEpisode })
		if ok {
			entries = []nyaa.Entry{*batchEntry}
		}
	}
	return episodeFilter(parseNyaaEntries(entries), latestTag, !useBatch)
}

func (c *Controller) nyaaFindAnime(ctx context.Context, entry animelist.Entry) ([]nyaa.Entry, error) {
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	strippedTitles := utils.ForEach(entry.Titles, parser.TitleStrip)
	titleQuery := nyaa.OrQuery(strippedTitles)
	sourceQuery := nyaa.OrQuery(c.dep.Config.Sources)
	qualityQuery := nyaa.OrQuery(c.dep.Config.Qualitites)
	entries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	log.Debug().Str("entry", entry.Titles[0]).Msgf("found %d torrents", len(entries))
	if err != nil {
		return nil, fmt.Errorf("getting nyaa list: %w", err)
	}
	// Filters only entries after the anime started airing.
	entries = utils.Filter[nyaa.Entry](entries, func(e nyaa.Entry) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, e.PubDate))
		return publishedDate.After(entry.StartDate)
	})
	return entries, nil
}

// DigestAnimeListEntry receives an anime list entry and fetches the anime feed, looking for new content.
func (c *Controller) DigestAnimeListEntry(ctx context.Context, entry animelist.Entry) (count int, err error) {
	nyaaEntries, err := c.nyaaFindAnime(ctx, entry)
	if err != nil {
		return count, err
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
	downloadFeed := filterNyaaFeed(nyaaEntries, latestTag, entry.AiringStatus)
	for _, nyaaEntry := range downloadFeed {
		if err := c.DigestNyaaTorrent(ctx, entry, nyaaEntry); err != nil {
			log.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
		count++
	}
	return count, nil
}
