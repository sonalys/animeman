package discovery

import (
	"context"
	"sync/atomic"
	"time"

	"github.com/expr-lang/expr/vm"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
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
		dep             Dependencies
		intervalTracker *intervalTracker
		lastRunErr      atomic.Pointer[error]

		// shokoQueue carries torrents added but not yet completed, consumed
		// by the shoko loop every minute until they are done.
		shokoQueue  chan parser.TorrentMetadata
		lastEntries []animelist.Entry
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep:             dep,
		intervalTracker: newIntervalTracker(dep.Config.PollFrequency),
		// Buffered so the discovery run never blocks on enqueue.
		shokoQueue: make(chan parser.TorrentMetadata, 100),
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

	if c.dep.Shoko != nil {
		// The shoko integration runs on its own schedule: torrents take a while
		// to download, so pending ones are retried every minute until completed.
		go c.runShokoLoop(ctx)
	}

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

// runShokoLoop retries pending torrents on a fixed one-minute interval,
// independent of the discovery scan cadence. Torrents are enqueued by the
// discovery run and stay queued until they complete in the torrent client.
func (c *Controller) runShokoLoop(ctx context.Context) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			if err := c.RunShokoIntegration(ctx, c.lastEntries); err != nil {
				log.Error().Msgf("shoko integration failed: %s", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

// enqueueShoko adds a torrent to the shoko queue, dropping it when the
// queue is full to avoid blocking the discovery run.
func (c *Controller) enqueueShoko(torrentMetadata parser.TorrentMetadata) {
	select {
	case c.shokoQueue <- torrentMetadata:
	default:
		log.Warn().
			Str("torrent", torrentMetadata.Torrent.Title).
			Msg("shoko queue full, dropping torrent")
	}
}
