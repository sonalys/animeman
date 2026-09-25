package controller

import (
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
		AnidbIDResolver   animelist.AnidbIDResolver
		AnidbEpIDResolver animelist.EpisodeAnidbIDResolver
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
		runner.Runner

		dep             Dependencies
		intervalTracker *intervalTracker
	}
)

func New(dep Dependencies) *Controller {
	return &Controller{
		dep:             dep,
		intervalTracker: newIntervalTracker(dep.Config.PollFrequency),
		Runner:          runner.New(),
	}
}

func (c *Controller) Start() {
	log.Info().Msgf("starting polling with frequency %s", c.dep.Config.PollFrequency.String())

	if c.dep.Shoko != nil && c.dep.AnidbEpIDResolver != nil && c.dep.AnidbIDResolver != nil {
		c.Runner.Run(c.startShokoRoutine)
	}

	c.Runner.Run(c.startDiscovery)
}
