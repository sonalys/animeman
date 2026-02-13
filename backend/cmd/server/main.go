package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/configs"
	"github.com/sonalys/animeman/internal/adapters/postgres"
	"github.com/sonalys/animeman/internal/utils/roundtripper"
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

	done()
}
