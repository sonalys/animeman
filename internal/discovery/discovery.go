package discovery

import (
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/parser"
)

type TaggedNyaa struct {
	meta             parser.ParsedTitle
	seasonEpisodeTag string
	entry            nyaa.Entry
}

func (c *Controller) RunDiscovery(ctx context.Context) error {
	t1 := time.Now()
	entries, err := c.dep.MAL.GetAnimeList(ctx,
		myanimelist.ListStatusWatching,
	)
	if err != nil {
		log.Fatal().Msgf("getting MAL list: %s", err)
	}
	log.Info().Msgf("processing %d entries from MAL", len(entries))
	var totalCount int
	for _, entry := range entries {
		count, err := c.DigestMALEntry(ctx, entry)
		if err != nil {
			if errors.Is(err, qbittorrent.ErrUnauthorized) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("failed to digest entry: %w", err)
			}
			continue
		}
		totalCount += count
	}
	if totalCount > 0 {
		log.Info().Msgf("added %d torrents", totalCount)
	}
	log.Info().Str("duration", time.Since(t1).String()).Msgf("discovery finished")
	return nil
}

func filterNyaaBatch(entries []nyaa.Entry) []nyaa.Entry {
	for _, entry := range entries {
		if meta := parser.ParseTitle(entry.Title); meta.IsMultiEpisode {
			return []nyaa.Entry{entry}
		}
	}
	return entries
}

func buildTaggedNyaaList(entries []nyaa.Entry) []TaggedNyaa {
	out := make([]TaggedNyaa, 0, len(entries))
	for _, entry := range entries {
		meta := parser.ParseTitle(entry.Title)
		out = append(out, TaggedNyaa{
			meta:             meta,
			seasonEpisodeTag: meta.BuildSeasonEpisodeTag(),
			entry:            entry,
		})
	}
	sort.Slice(out, func(i, j int) bool {
		return compareTags(out[i].seasonEpisodeTag, out[j].seasonEpisodeTag) < 0
	})
	return out
}

func filterEpisodes(list []TaggedNyaa, latestTag string) []TaggedNyaa {
	out := make([]TaggedNyaa, 0, len(list))
	for _, nyaaEntry := range list {
		// Make sure we only add episodes ahead of the current ones in the qBittorrent.
		if compareTags(nyaaEntry.seasonEpisodeTag, latestTag) <= 0 {
			continue
		}
		latestTag = nyaaEntry.seasonEpisodeTag
		out = append(out, nyaaEntry)
	}
	return out
}

func (c *Controller) DigestMALEntry(ctx context.Context, entry myanimelist.AnimeListEntry) (count int, err error) {
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	titleQuery := nyaa.OrQuery{parser.StripTitle(entry.TitleEng), parser.StripTitle(entry.Title)}
	sourceQuery := nyaa.OrQuery(c.dep.Config.Sources)
	qualityQuery := nyaa.OrQuery(c.dep.Config.Qualitites)

	nyaaEntries, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	log.Debug().Str("entry", entry.GetTitle()).Msgf("found %d torrents", len(nyaaEntries))
	if err != nil {
		return count, fmt.Errorf("getting nyaa list: %w", err)
	}
	// There should always be torrents for entries, if there aren't we can just exit the routine.
	if len(nyaaEntries) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.GetTitle())
		return count, nil
	}
	latestTag, err := c.GetLatestTag(ctx, entry)
	if err != nil {
		return count, fmt.Errorf("getting latest tag: %w", err)
	}
	// If we don't have any episodes, and show is released, try to find a batch for all episodes.
	if latestTag == "" && entry.AiringStatus == myanimelist.AiringStatusAired {
		nyaaEntries = filterNyaaBatch(nyaaEntries)
	}
	taggedNyaaList := buildTaggedNyaaList(nyaaEntries)
	taggedNyaaList = filterEpisodes(taggedNyaaList, latestTag)
	for _, nyaaEntry := range taggedNyaaList {
		if err := c.DigestNyaaTorrent(ctx, entry, nyaaEntry); err != nil {
			log.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
		count++
	}
	return count, nil
}
