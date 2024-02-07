package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/config"
	"github.com/sonalys/animeman/internal/discovery"
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
		MAL:  myanimelist.New(config.MALUser),
		NYAA: nyaa.New(),
		QB:   qbittorrent.New(ctx, config.QBitTorrentHost, config.QBitTorrentUsername, config.QBitTorrentPassword),
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
		log.Error().Msgf("failed to finish discover: %s", err)
	} else {
		log.Info().Msg("finished successfully")
	}
	done()
}
