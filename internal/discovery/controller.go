package discovery

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/integrations/nyaa"
)

type (
	Dependencies struct {
		NYAA            *nyaa.API
		AnimeListClient AnimeListSource
		TorrentClient   TorrentClient
		Config          Config
	}

	Controller struct {
		dep             Dependencies
		intervalTracker *IntervalTracker
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep:             dep,
		intervalTracker: NewIntervalTracker(dep.Config.PollFrequency),
	}
}

func (c *Controller) Start(ctx context.Context) error {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())

	ticker := time.NewTicker(c.dep.Config.PollFrequency)
	defer ticker.Stop()

	for {
		if err := c.RunDiscovery(ctx); err != nil {
			log.Error().Msgf("discovery scan failed: %s", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			log.Info().Msgf("stopping discovery: %s", ctx.Err())
			return nil
		}
	}
}
