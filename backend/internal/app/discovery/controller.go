package discovery

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/nyaa"
)

type (
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
			if err := c.RunDiscovery(ctx); err != nil {
				log.Error().Msgf("scan failed: %s", err)
			}
		case <-ctx.Done():
			return nil
		}
	}
}
