package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/anilist"
	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/integrations/nyaa"
	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/configs"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/utils"
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func initializeAnimeList(c configs.AnimeListConfig) discovery.AnimeListSource {
	switch c.Type {
	case configs.AnimeListTypeMAL:
		return myanimelist.New(c.Username)
	case configs.AnimeListTypeAnilist:
		return anilist.New(c.Username)
	default:
		log.Panic().Msgf("animeListType %s not implemented", c.Type)
	}
	return nil
}

func initializeTorrentClient(ctx context.Context, c configs.TorrentConfig) discovery.TorrentClient {
	switch c.Type {
	case configs.TorrentClientTypeQBittorrent:
		return qbittorrent.New(ctx, c.Host, c.Username, c.Password)
	default:
		log.Panic().Msgf("animeListType %s not implemented", c.Type)
	}
	return nil
}

func main() {
	log.Info().Msg("starting Animeman")
	config, err := configs.ReadConfig(utils.Coalesce(os.Getenv("CONFIG_PATH"), "config.yaml"))
	if err != nil {
		log.Fatal().Msgf("config is not valid: %s", err)
	}
	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	c := discovery.New(discovery.Dependencies{
		NYAA: nyaa.New(nyaa.Config{
			ListParameters: config.CustomParameters,
		}),
		AnimeListClient: initializeAnimeList(config.AnimeListConfig),
		TorrentClient:   initializeTorrentClient(ctx, config.TorrentConfig),
		Config: discovery.Config{
			Sources:          config.Sources,
			Qualitites:       config.Qualities,
			Category:         config.Category,
			RenameTorrent:    *utils.Coalesce(config.RenameTorrent, utils.Pointer(true)),
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
