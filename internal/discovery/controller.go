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

type (
	Config struct {
		Sources          []string
		Qualitites       []string
		Category         string
		DownloadPath     string
		CreateShowFolder bool
		PollFrequency    time.Duration
	}

	Dependencies struct {
		MAL    *myanimelist.API
		NYAA   *nyaa.API
		QB     *qbittorrent.API
		Config Config
	}

	Controller struct {
		dep Dependencies
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep: dep,
	}
}

func (c *Controller) Start(ctx context.Context) error {
	if err := c.UpdateExistingTorrentsTags(ctx); err != nil {
		return fmt.Errorf("updating qBittorrent entries: %w", err)
	}
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())
	timer := time.NewTicker(c.dep.Config.PollFrequency)
	defer timer.Stop()
	c.RunDiscovery(ctx)
	for {
		select {
		case <-timer.C:
			err := c.RunDiscovery(ctx)
			if err == nil {
				continue
			}
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Error().Msgf("scan failed: %s", err)
		case <-ctx.Done():
			return nil
		}
	}
}

func buildSeasonEpisodeTag(parsedTitle parser.ParsedTitle) string {
	resp := fmt.Sprintf("%s S%s", parsedTitle.Title, parsedTitle.Season)
	if !parsedTitle.IsMultiEpisode {
		resp = resp + fmt.Sprintf("E%s", parsedTitle.Episode)
	}
	return resp
}

func buildTorrentTags(parsedTitle parser.ParsedTitle) qbittorrent.Tags {
	tags := qbittorrent.Tags{"!" + parsedTitle.Title, buildSeasonEpisodeTag(parsedTitle)}
	if parsedTitle.IsMultiEpisode {
		tags = append(tags, buildBatchTag(parsedTitle))
	}
	return tags
}

func (c *Controller) UpdateExistingTorrentsTags(ctx context.Context) error {
	torrents, err := c.dep.QB.List(ctx, qbittorrent.Category(c.dep.Config.Category))
	if err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for _, torrent := range torrents {
		parsedTitle := parser.ParseTitle(torrent.Name)
		if err := c.dep.QB.AddTorrentTags(ctx, []string{torrent.Hash}, buildTorrentTags(parsedTitle)); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}
	return nil
}

func (c *Controller) RunDiscovery(ctx context.Context) error {
	log.Info().Msg("discovery started")
	entries, err := c.dep.MAL.GetAnimeList(ctx,
		myanimelist.ListStatusWatching,
	)
	if err != nil {
		log.Fatal().Msgf("getting MAL list: %s", err)
	}
	log.Info().Msgf("processing %d entries from MAL", len(entries))
	var totalCount int
	for _, entry := range entries {
		count, err := c.digestEntry(ctx, entry)
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
	return nil
}

func buildBatchTag(parsedTitle parser.ParsedTitle) string {
	return fmt.Sprintf("%s S%s batch", parsedTitle.Title, parsedTitle.Season)
}

func (c *Controller) doesConflict(ctx context.Context, parsedTitle parser.ParsedTitle) (bool, error) {
	// check if torrent already exists, if so we skip it.
	torrentList, _ := c.dep.QB.List(ctx, qbittorrent.Tag(buildBatchTag(parsedTitle)))
	if len(torrentList) > 0 {
		return true, nil
	}
	torrentList, _ = c.dep.QB.List(ctx, qbittorrent.Tag(buildSeasonEpisodeTag(parsedTitle)))
	if len(torrentList) > 0 {
		return true, nil
	}
	return false, nil
}

func (c *Controller) digestEntry(ctx context.Context, entry myanimelist.AnimeListEntry) (count int, err error) {
	log.Debug().Msgf("Digesting entry '%s'", entry.GetTitle())
	torrents, err := c.dep.NYAA.List(ctx,
		nyaa.CategoryAnimeEnglishTranslated,
		nyaa.OrQuery{parser.StripTitle(entry.TitleEng), parser.StripTitle(entry.Title)},
		nyaa.OrQuery(c.dep.Config.Sources),
		nyaa.OrQuery(c.dep.Config.Qualitites),
	)
	log.Debug().Str("entry", entry.GetTitle()).Msgf("Found %d torrents", len(torrents))
	if err != nil {
		return 0, fmt.Errorf("getting nyaa list: %w", err)
	}
	if len(torrents) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.GetTitle())
		return 0, nil
	}
	for i := range torrents {
		torrent := torrents[i]
		log.Debug().Str("entry", entry.GetTitle()).Msgf("analyzing torrent '%s'", torrent.Title)
		meta := parser.ParseTitle(torrent.Title)
		if meta.IsMultiEpisode && entry.AiringStatus == myanimelist.AiringStatusAiring {
			log.Debug().Msgf("torrent dropped: multi-episode for currently airing")
			continue
		}
		doesConflict, err := c.doesConflict(ctx, meta)
		if err != nil || doesConflict {
			log.Debug().Str("tags", buildSeasonEpisodeTag(meta)).Msgf("torrent is conflicting")
			break
		}
		var savePath qbittorrent.SavePath
		if c.dep.Config.CreateShowFolder {
			savePath = qbittorrent.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
		} else {
			savePath = qbittorrent.SavePath(c.dep.Config.DownloadPath)
		}
		tags := buildTorrentTags(meta)
		err = c.dep.QB.AddTorrent(ctx,
			tags,
			savePath,
			qbittorrent.TorrentURL{torrent.Link},
			qbittorrent.Category(c.dep.Config.Category),
		)
		if err != nil {
			return count, fmt.Errorf("adding torrents: %w", err)
		}
		log.Info().
			Str("savePath", string(savePath)).
			Strs("tag", tags).
			Msgf("torrent added")
	}
	return count, nil
}
