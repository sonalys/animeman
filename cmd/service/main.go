package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/configs"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/integrations/anilist"
	"github.com/sonalys/animeman/internal/integrations/myanimelist"
	"github.com/sonalys/animeman/internal/integrations/nyaa"
	"github.com/sonalys/animeman/internal/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/roundtripper"
	"github.com/sonalys/animeman/internal/utils"
	"golang.org/x/time/rate"
)

const (
	userAgent = "github.com/sonalys/animeman"
)

var (
	version          = "development"
	defaultTransport = roundtripper.NewUserAgentTransport(
		roundtripper.NewLoggerTransport(http.DefaultTransport),
		userAgent,
	)
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	zerolog.DefaultContextLogger = &log.Logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func initializeAnimeList(c configs.AnimeListConfig) discovery.AnimeListSource {
	httpClient := &http.Client{
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Second), 1),
		),
		Timeout: 15 * time.Second,
	}

	switch c.Type {
	case configs.AnimeListTypeMAL:
		return myanimelist.New(httpClient, c.Username)
	case configs.AnimeListTypeAnilist:
		return anilist.New(httpClient, c.Username)
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
	log.Info().Msgf("starting Animeman [%s]", version)

	config, err := configs.ReadConfig(utils.Coalesce(os.Getenv("CONFIG_PATH"), "config.yaml"))
	if err != nil {
		log.Fatal().Msgf("config is not valid: %s", err)
	}

	zerolog.SetGlobalLevel(config.LogLevel.Convert())

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	nyaaClient := &http.Client{
		Jar: http.DefaultClient.Jar,
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Second), 1),
		),
		Timeout: 15 * time.Second,
	}

	nyaaConfig := nyaa.Config{
		ListParameters: config.CustomParameters,
	}

	c := discovery.New(discovery.Dependencies{
		NYAA:            nyaa.New(nyaaClient, nyaaConfig),
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
