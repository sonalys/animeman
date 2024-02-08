package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/internal/config"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/myanimelist"
	"github.com/sonalys/animeman/internal/qbittorrent"
	"github.com/sonalys/animeman/internal/utils"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	log.Info().Msg("starting Animeman")
	config := config.ReadConfig(utils.Coalesce(os.Getenv("CONFIG_PATH"), "config.yaml"))
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	c := discovery.New(discovery.Dependencies{
		NYAA:            nyaa.New(),
		AnimeListClient: myanimelist.New(config.AnimeListConfig.Username),
		TorrentClient:   qbittorrent.New(ctx, config.TorrentConfig.Host, config.TorrentConfig.Username, config.TorrentConfig.Password),
		Config: discovery.Config{
			Sources:          config.Sources,
			Qualitites:       config.Qualities,
			Category:         config.Category,
			DownloadPath:     config.DownloadPath,
			CreateShowFolder: config.CreateShowFolder,
			PollFrequency:    config.PollFrequency,
		},
	})
	if err := c.Start(ctx); err != nil {
		log.Error().Msgf("failed to shutdown: %s", err)
	} else {
		log.Info().Msg("shutdown successful")
	}
	done()
}
