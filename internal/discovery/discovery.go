package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/parser"
)

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

func (c *Controller) DigestMALEntry(ctx context.Context, entry myanimelist.AnimeListEntry) (count int, err error) {
	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	titleQuery := nyaa.OrQuery{parser.StripTitle(entry.TitleEng), parser.StripTitle(entry.Title)}
	sourceQuery := nyaa.OrQuery(c.dep.Config.Sources)
	qualityQuery := nyaa.OrQuery(c.dep.Config.Qualitites)

	torrents, err := c.dep.NYAA.List(ctx, titleQuery, sourceQuery, qualityQuery)
	log.Debug().Str("entry", entry.GetTitle()).Msgf("found %d torrents", len(torrents))
	if err != nil {
		return 0, fmt.Errorf("getting nyaa list: %w", err)
	}
	// There should always be torrents for entries, if there aren't we can just exit the routine.
	if len(torrents) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.GetTitle())
		return 0, nil
	}
	for _, torrent := range torrents {
		log.Debug().Str("entry", entry.GetTitle()).Msgf("analyzing torrent '%s'", torrent.Title)
		added, err := c.DigestNyaaTorrent(ctx, entry, torrent)
		if err != nil {
			log.Error().Msgf("failed to digest nyaa entry: %s", err)
			continue
		}
		if added {
			count++
		}
	}
	return count, nil
}
