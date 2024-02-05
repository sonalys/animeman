package controller

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

type Config struct {
	Sources          []string
	Qualitites       []string
	Category         string
	DownloadPath     string
	CreateShowFolder bool
	PollFrequency    time.Duration
}

type Dependencies struct {
	MAL    *myanimelist.API
	NYAA   *nyaa.API
	QB     *qbittorrent.API
	Config Config
}

type Controller struct {
	dep Dependencies
}

func New(dep Dependencies) *Controller {
	return &Controller{
		dep: dep,
	}
}

func (c *Controller) Start(ctx context.Context) {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())
	timer := time.NewTicker(c.dep.Config.PollFrequency)
	defer timer.Stop()
	c.scan(ctx)
	for {
		select {
		case <-timer.C:
			c.scan(ctx)
		case <-ctx.Done():
			return
		}
	}
}

func (c *Controller) scan(ctx context.Context) {
	log.Info().Msg("discovery started")
	entries, err := c.dep.MAL.GetAnimeList(ctx, myanimelist.ListStatusWatching)
	if err != nil {
		panic(fmt.Errorf("getting MAL list: %w", err))
	}
	log.Info().Msgf("processing %d entries from MAL", len(entries))
	var addedCount int
	for _, entry := range entries {
		torrents, err := c.dep.NYAA.List(ctx,
			nyaa.CategoryAnimeEnglishTranslated,
			nyaa.Query(parser.StripTitle(entry.GetTitle())),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Sources, "|"))),
			nyaa.Query(fmt.Sprintf("(%s)", strings.Join(c.dep.Config.Qualitites, "|"))),
			nyaa.Query("(^~|batch)"),
		)
		if err != nil {
			log.Error().Msgf("getting nyaa list: %s", err)
			break
		}
		added, err := c.digestEntry(ctx, entry, torrents)
		if err != nil {
			log.Error().Msgf("failed to digest entry: %s\n", err)
			if errors.Is(err, qbittorrent.ErrUnauthorized) {
				break
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
}

func (c *Controller) digestEntry(ctx context.Context, entry myanimelist.AnimeListEntry, torrents []nyaa.Entry) (bool, error) {
	if len(torrents) == 0 {
		return false, nil
	}
	// We only add the most recent entry for now.
	torrent := torrents[0]
	parsedTitle := parser.ParseTitle(torrent.Title)
	tags := qbittorrent.Tags{"animeman", entry.GetTitle(), fmt.Sprintf("S%sE%s", parsedTitle.Season, parsedTitle.Episode)}
	// check if torrent already exists, if so we skip it.
	torrentList, err := c.dep.QB.List(ctx, tags)
	if err != nil {
		return false, fmt.Errorf("listing torrents: %w", err)
	}
	if len(torrentList) > 0 {
		return false, nil
	}
	var savePath qbittorrent.SavePath
	if c.dep.Config.CreateShowFolder {
		savePath = qbittorrent.SavePath(fmt.Sprintf("%s/%s", c.dep.Config.DownloadPath, entry.GetTitle()))
	} else {
		savePath = qbittorrent.SavePath(c.dep.Config.DownloadPath)
	}
	err = c.dep.QB.AddTorrent(ctx,
		qbittorrent.TorrentURL{torrent.Link},
		tags,
		qbittorrent.Category(c.dep.Config.Category),
		savePath,
	)
	if err != nil {
		return false, fmt.Errorf("adding torrents: %w", err)
	}
	log.Info().
		Str("savePath", string(savePath)).
		Msgf("torrent '%s' added", entry.GetTitle())
	return true, nil
}
