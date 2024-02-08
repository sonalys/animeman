package discovery

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
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

	AnimeListSource interface {
		GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error)
	}

	TorrentClient interface {
		List(ctx context.Context, args ...torrentclient.ArgListTorrent) ([]torrentclient.Torrent, error)
		AddTorrent(ctx context.Context, args ...torrentclient.ArgAddTorrent) error
		AddTorrentTags(ctx context.Context, hashes []string, args ...torrentclient.AddTorrentTagsArg) error
	}

	Dependencies struct {
		NYAA            *nyaa.API
		AnimeListClient AnimeListSource
		TorrentClient   TorrentClient
		Config          Config
	}

	Controller struct {
		dep Dependencies
	}
)

func New(dep Dependencies) *Controller {
	if dep.Config.PollFrequency < 10*time.Second {
		log.Fatal().Msgf("pollFrequency cannot be less than 10 seconds. was %s", dep.Config.PollFrequency)
	}
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
	if err := c.RunDiscovery(ctx); err != nil {
		log.Error().Msgf("discovery failed: %s", err)
	}
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
