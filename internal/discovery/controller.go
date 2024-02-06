package discovery

import (
	"context"
	"errors"
	"fmt"
	"strings"
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
		return fmt.Errorf("updating qBitTorrent entries: %w", err)
	}
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())
	timer := time.NewTicker(c.dep.Config.PollFrequency)
	defer timer.Stop()
	c.RunDiscovery(ctx)
	for {
		select {
		case <-timer.C:
			err := c.RunDiscovery(ctx)
			if errors.Is(err, context.Canceled) {
				return err
			}
			log.Error().Msgf("scan failed: %s", err)
		case <-ctx.Done():
			return nil
		}
	}
}

func buildTorrentTags(title string) qbittorrent.Tags {
	parsedTitle := parser.ParseTitle(title)
	tags := qbittorrent.Tags{"animeman", parsedTitle.Title, buildSeasonEpisodeTag(parsedTitle.Season, parsedTitle.Episode)}
	if parsedTitle.Episode == "0" && parsedTitle.IsMultiEpisode {
		tags = append(tags, TagBatch)
	}
	return tags
}

func (c *Controller) UpdateExistingTorrentsTags(ctx context.Context) error {
	torrents, err := c.dep.QB.List(ctx, qbittorrent.Category(c.dep.Config.Category))
	if err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for i := range torrents {
		torrent := &torrents[i]
		if err := c.dep.QB.RemoveTorrentTags(ctx, []string{torrent.Hash}); err != nil {
			log.Warn().Msgf("failed to remove tags from torrent '%s'", torrent.Name)
		}
		if err := c.dep.QB.AddTorrentTags(ctx, []string{torrent.Hash}, buildTorrentTags(torrent.Name)); err != nil {
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
	var addedCount int
	for _, entry := range entries {
		log.Debug().Msgf("Digesting entry '%s'", entry.GetTitle())
		torrents, err := c.dep.NYAA.List(ctx,
			nyaa.CategoryAnimeEnglishTranslated,
			nyaa.Query(parser.StripTitle(entry.GetTitle())),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Sources, "|"))),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Qualitites, "|"))),
		)
		log.Debug().Str("entry", entry.GetTitle()).Msgf("Found %d torrents", len(torrents))
		if err != nil {
			return fmt.Errorf("getting nyaa list: %w", err)
		}
		added, err := c.digestEntry(ctx, entry, torrents)
		if err != nil {
			if errors.Is(err, qbittorrent.ErrUnauthorized) || errors.Is(err, context.Canceled) {
				return fmt.Errorf("failed to digest entry: %w", err)
			}
			continue
		}
		if added {
			addedCount++
		}
	}
	if addedCount > 0 {
		log.Info().Msgf("added %d torrents", addedCount)
	}
	return nil
}

func buildSeasonEpisodeTag(season, episode string) string {
	resp := ""
	if season != "0" {
		resp = fmt.Sprintf("S%s", season)
	}
	if episode != "0" {
		resp = resp + fmt.Sprintf("E%s", episode)
	}
	return resp
}

var TagBatch = "batch"

func (c *Controller) doesConflict(ctx context.Context, title string, parsedTitle parser.ParsedTitle) (bool, error) {
	// check if torrent already exists, if so we skip it.
	torrentList, err := c.dep.QB.List(ctx, qbittorrent.Tags{
		title, TagBatch,
	})
	if err != nil {
		return false, fmt.Errorf("listing torrents: %w", err)
	}
	if len(torrentList) > 0 {
		return true, nil
	}
	torrentList, err = c.dep.QB.List(ctx, qbittorrent.Tags{
		title, buildSeasonEpisodeTag(parsedTitle.Season, parsedTitle.Episode),
	})
	if err != nil {
		return false, fmt.Errorf("listing torrents: %w", err)
	}
	if len(torrentList) > 0 {
		return true, nil
	}
	return false, nil
}

func (c *Controller) digestEntry(ctx context.Context, entry myanimelist.AnimeListEntry, torrents []nyaa.Entry) (bool, error) {
	if len(torrents) == 0 {
		log.Error().Msgf("no torrents found for entry '%s'", entry.GetTitle())
		return false, nil
	}
	var torrent nyaa.Entry
	var found bool
	for i := range torrents {
		torrent = torrents[i]
		log.Debug().Str("entry", entry.GetTitle()).Msgf("Analyzing torrent '%s'", torrent.Title)
		parsedTitle := parser.ParseTitle(torrent.Title)
		if parsedTitle.IsMultiEpisode && entry.AiringStatus == myanimelist.AiringStatusAiring {
			log.Debug().Str("entry", entry.GetTitle()).Msgf("torrent '%s' dropped: multi-episode for currently airing", torrent.Title)
			continue
		}
		doesConflict, err := c.doesConflict(ctx, entry.GetTitle(), parsedTitle)
		if err != nil || !doesConflict {
			found = true
			break
		}
	}
	if !found {
		return false, nil
	}
	var savePath qbittorrent.SavePath
	if c.dep.Config.CreateShowFolder {
		savePath = qbittorrent.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
	} else {
		savePath = qbittorrent.SavePath(c.dep.Config.DownloadPath)
	}
	err := c.dep.QB.AddTorrent(ctx,
		buildTorrentTags(torrent.Title),
		savePath,
		qbittorrent.TorrentURL{torrent.Link},
		qbittorrent.Category(c.dep.Config.Category),
		qbittorrent.Paused(true),
	)
	if err != nil {
		return false, fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().
		Str("savePath", string(savePath)).
		Msgf("torrent '%s' added", entry.GetTitle())
	return true, nil
}
