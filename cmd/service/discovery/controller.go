package discovery

import (
	"context"
	"errors"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/pkg/runner"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/shoko"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

type (
	Dependencies struct {
		TorrentSource     torrentsource.Source
		AnimeListSource   animelist.AnimeListSource
		TorrentClient     torrentclient.Client
		AnilistIDResolver animelist.AnilistIDResolver
		// Shoko is optional, nil disables the shoko integration.
		Shoko  shoko.Shoko
		Config Config
	}

	Config struct {
		SearchSuffix     string
		ReleaseGroups    []string
		Qualitites       []string
		Category         string
		RenameTorrent    bool
		RenameFormat     *vm.Program
		DownloadPath     string
		CreateShowFolder bool
		PollFrequency    time.Duration
	}

	Controller struct {
		dep                           Dependencies
		intervalTracker               *intervalTracker
		lastShokoUnknownFileTimestamp time.Time
		runner                        runner.Runner
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep:             dep,
		intervalTracker: newIntervalTracker(dep.Config.PollFrequency),
		runner:          runner.New(),
	}
}

func (c *Controller) Start() {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())

	if c.dep.Shoko != nil {
		c.runner.Run(c.runShokoLoop)
	}

	c.runner.Run(c.runDiscoveryLoop)
}

func (c *Controller) Shutdown(ctx context.Context) error {
	return c.runner.Shutdown(ctx)
}

func (c *Controller) runDiscoveryLoop(ctx context.Context, shutdown <-chan struct{}) {
	ticker := time.NewTicker(c.dep.Config.PollFrequency)
	defer ticker.Stop()

	for {
		if err := c.RunDiscovery(ctx); err != nil {
			log.Error().Msgf("discovery scan failed: %s", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		case <-shutdown:
			return
		}
	}
}

func (c *Controller) runShokoLoop(ctx context.Context, shutdown <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		if err := c.RunShokoIntegration(ctx); err != nil {
			if errors.Is(err, shoko.ErrForbidden) {
				log.Error().Msg("shoko is unauthorized, routine will stop")
				return
			}
			log.Error().Msgf("shoko integration failed: %s", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		case <-shutdown:
			return
		}
	}
}
