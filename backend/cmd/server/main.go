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
	"github.com/sonalys/animeman/cmd/server/handler"
	indexingclient "github.com/sonalys/animeman/internal/adapters/indexing"
	"github.com/sonalys/animeman/internal/adapters/transferclient"
	"github.com/sonalys/animeman/internal/app/jwt"
	"github.com/sonalys/animeman/internal/app/monitoring"
	"github.com/sonalys/animeman/internal/app/usecases"
	"github.com/sonalys/animeman/internal/utils/optional"
	"github.com/sonalys/animeman/internal/utils/otel"
	"github.com/sonalys/animeman/internal/utils/roundtripper"
)

var (
	version = "dev"

	_ = roundtripper.NewUserAgentTransport(
		roundtripper.NewLoggerTransport(http.DefaultTransport),
		"github.com/sonalys/animeman@"+version,
	)
)

func init() {
	log.Logger = log.
		Hook(otel.OTelHook{}).
		Output(zerolog.ConsoleWriter{
			Out:        os.Stdout,
			TimeFormat: time.DateTime,
		})

	zerolog.LevelColors[zerolog.DebugLevel] = 35
	zerolog.DefaultContextLogger = &log.Logger
}

func main() {
	log.Info().
		Str("version", version).
		Msgf("Starting Animeman")

	logLvString := optional.Coalesce(os.Getenv("LOG_LEVEL"), "info")
	logLv, err := zerolog.ParseLevel(logLvString)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not parse log level")
	}

	postgresConnStr := os.Getenv("POSTGRES_CONN")
	if postgresConnStr == "" {
		log.Fatal().
			Err(err).
			Msg("Missing env POSTGRES_CONN")
	}

	log.Info().Stringer("logLevel", logLv).Msg("Adjusting log verbosity")
	zerolog.SetGlobalLevel(logLv)

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)

	telemetryShutdown, err := otel.Initialize(ctx, "jaeger:4317", version)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not initialize telemetry")
	}

	log.Info().
		Msg("Connecting to PostgreSQL")

	adapters := initializeAdapters(ctx, postgresConnStr)
	jwtClient := jwt.NewClient([]byte("secret"))

	repositories := usecases.Repositories{
		UserRepository:           adapters.postgresClient.UserRepository(),
		IndexerClientRepository:  adapters.postgresClient.IndexerClientRepository(),
		TransferClientRepository: adapters.postgresClient.TransferClientRepository(),
		CollectionRepository:     adapters.postgresClient.CollectionRepository(),
		WatchlistRepository:      adapters.postgresClient.WatchlistRepository(),
	}

	factories := usecases.Factories{
		TransferClientControllerFactory: transferclient.NewFactory(),
		IndexingClientControllerFactory: indexingclient.NewFactory(),
	}

	watcher := monitoring.New(repositories.CollectionRepository)

	go func() {
		log.Info().
			Msg("Starting collection watcher")

		if err := watcher.Start(ctx); err != nil {
			log.Fatal().
				Err(err).
				Msg("Could not initialize collection watcher")
		}
	}()

	usecases := usecases.NewUsecases(repositories, factories)

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

	log.Info().
		Str("addr", httpServer.Addr).
		Msg("Serving http api")

	if err := httpServer.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Fatal().
			Err(err).
			Msg("Could not serve http api")
	}

	log.Info().
		Msg("Goodbye!")

	done()
}
