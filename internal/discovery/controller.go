package discovery

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/shoko"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

type (
	Dependencies struct {
		TorrentSource     torrentsource.Source
		AnimeListSource   animelist.AnimeListSource
		TorrentClient     torrentclient.TorrentClient
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
		dep             Dependencies
		intervalTracker *intervalTracker
		lastRunErr      atomic.Pointer[error]
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep:             dep,
		intervalTracker: newIntervalTracker(dep.Config.PollFrequency),
	}
}

// LastRunErr returns the error from the last discovery scan, nil when the last scan succeeded or no scan has run yet.
func (c *Controller) LastRunErr() error {
	p := c.lastRunErr.Load()
	if p == nil {
		return nil
	}
	return *p
}

// Deps exposes the controller dependencies for health checks.
func (c *Controller) Deps() Dependencies {
	return c.dep
}

func (c *Controller) Start(ctx context.Context) error {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())

	ticker := time.NewTicker(c.dep.Config.PollFrequency)
	defer ticker.Stop()

	for {
		err := c.RunDiscovery(ctx)
		c.lastRunErr.Store(&err)
		if err != nil {
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
