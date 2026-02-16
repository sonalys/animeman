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
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/usecases"
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
		Out: os.Stdout,
	}).Hook(
		otel.OTelHook{},
	)

	zerolog.DefaultContextLogger = &log.Logger
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
			Str("path", configPath).
			Msg("Invalid configuration")
	}

	newLogLevel := config.LogLevel.Convert()
	log.Info().Stringer("logLevel", newLogLevel).Msg("Adjusting log verbosity")
	zerolog.SetGlobalLevel(config.LogLevel.Convert())

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	telemetryShutdown, err := otel.Initialize(ctx, "jaeger:4317", version)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not initialize telemetry")
	}

	log.Info().
		Msg("Connecting to PostgreSQL")

	adapters := initializeAdapters(ctx, config)
	jwtClient := jwt.NewClient([]byte("secret"))
	usecases := usecases.NewUsecases(usecases.Repositories{
		UserRepository: adapters.postgresClient.UserRepository(),
	})

	handler, err := handler.New(jwtClient, usecases)
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

	go func() {
		<-ctx.Done()

		ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
		defer cancel()

		go func() {
			if err := httpServer.Shutdown(ctx); err != nil {
				log.Error().
					Err(err).
					Msg("Failed to shutdown http server gracefully")
			}
		}()

		go func() {
			if err := telemetryShutdown(ctx); err != nil {
				log.Error().
					Err(err).
					Msg("Failed to shutdown the telemetry gracefully")
			}
		}()
	}()

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().
			Err(err).
			Msg("Could not serve http api")
	}

	log.Info().
		Msg("Goodbye!")

	done()
}
