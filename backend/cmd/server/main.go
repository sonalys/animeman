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
	"github.com/sonalys/animeman/cmd/server/configs"
	"github.com/sonalys/animeman/internal/adapters/nyaa"
	"github.com/sonalys/animeman/internal/adapters/postgres"
	"github.com/sonalys/animeman/internal/app/discovery"
	"github.com/sonalys/animeman/internal/utils/optional"
	"github.com/sonalys/animeman/internal/utils/roundtripper"
	"golang.org/x/time/rate"
)

var (
	version          = "dev"
	defaultTransport = roundtripper.NewUserAgentTransport(
		roundtripper.NewLoggerTransport(http.DefaultTransport),
		"github.com/sonalys/animeman",
	)
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	})

	zerolog.DefaultContextLogger = &log.Logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func main() {
	log.Info().
		Str("version", version).
		Msgf("Starting Animeman")

	configPath := configs.ReadEnv("CONFIG_PATH", "config.yaml")

	config, err := configs.ReadConfig(configPath)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Invalid configuration")
	}

	zerolog.SetGlobalLevel(config.LogLevel.Convert())

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	log.Info().
		Msg("Connecting to PostgreSQL")

	_, err = postgres.New(ctx, config.PostgresConfig.ConnStr)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not connect to PostgreSQL")
	}

	nyaaHTTPClient := &http.Client{
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
		NYAA:            nyaa.New(nyaaHTTPClient, nyaaConfig),
		AnimeListClient: initializeAnimeList(config.AnimeListConfig),
		TorrentClient:   initializeTorrentClient(ctx, config.TorrentConfig),
		Config: discovery.Config{
			Sources:          config.Sources,
			Qualitites:       config.Qualities,
			Category:         config.Category,
			RenameTorrent:    *optional.Coalesce(config.RenameTorrent, new(true)),
			DownloadPath:     config.DownloadPath,
			CreateShowFolder: config.CreateShowFolder,
			PollFrequency:    config.PollFrequency,
		},
	})
	if err := c.Start(ctx); err != nil {
		log.Info().
			Err(err).
			Msg("Failed to shutdown properly")
	} else {
		log.Info().Msg("Goodbye!")
	}
	done()
}
