package discovery

import (
	"context"
	"errors"
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
		RenameTorrent    bool
		DownloadPath     string
		CreateShowFolder bool
		PollFrequency    time.Duration
	}

	AnimeListSource interface {
		GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error)
	}

	TorrentClient interface {
		List(ctx context.Context, arg *torrentclient.ListTorrentConfig) ([]torrentclient.Torrent, error)
		AddTorrent(ctx context.Context, arg *torrentclient.AddTorrentConfig) error
		AddTorrentTags(ctx context.Context, hashes []string, tags []string) error
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
	return &Controller{
		dep: dep,
	}
}

func (c *Controller) Start(ctx context.Context) error {
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
