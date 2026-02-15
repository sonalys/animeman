package main

import (
	"context"
	"errors"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/configs"
	"github.com/sonalys/animeman/cmd/server/handler"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/roundtripper"
)

var (
	version = "dev"

	_ = roundtripper.NewUserAgentTransport(
		roundtripper.NewLoggerTransport(http.DefaultTransport),
		"github.com/sonalys/animeman",
	)
)

func init() {
	log.Logger = log.Output(zerolog.ConsoleWriter{
		Out: os.Stderr,
	}).Hook(
		otel.OTelHook{},
	)

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

	_ = initializeAdapters(ctx, config)

	handler, err := handler.New()
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not initialize http handler")
	}

	httpServer := http.Server{
		Addr:         ":8080",
		WriteTimeout: 15 * time.Second,
		ReadTimeout:  10 * time.Second,
		Handler:      handler,
	}

	log.Info().
		Str("addr", httpServer.Addr).
		Msg("Serving http api")

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().
			Err(err).
			Msg("Could not serve http api")
	}

	done()
}
