package main

import (
	"context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/expr-lang/expr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/service/controller"
	"github.com/sonalys/animeman/internal/adapters/animelist/anilist"
	"github.com/sonalys/animeman/internal/adapters/animelist/myanimelist"
	shokoadapter "github.com/sonalys/animeman/internal/adapters/shoko"
	"github.com/sonalys/animeman/internal/adapters/torrentclient/qbittorrent"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nekobt"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nyaa"
	"github.com/sonalys/animeman/internal/pkg/coalesce"
	"github.com/sonalys/animeman/internal/pkg/http/roundtripper"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/shoko"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"golang.org/x/sync/errgroup"
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
	writer := zerolog.ConsoleWriter{
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	}

	log.Logger = log.
		Output(writer).
		With().
		Caller().
		Logger()

	zerolog.DefaultContextLogger = &log.Logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func initializeAnimeList(
	httpClient *http.Client,
	config AnimeListConfig,
	anilistAPI *anilist.API,
) animelist.AnimeListSource {
	switch config.Type {
	case AnimeListTypeMAL:
		return myanimelist.New(httpClient, config.Username, config.CacheTTL)
	case AnimeListTypeAnilist:
		return anilistAPI
	default:
		log.Panic().Msgf("animeListType %s not implemented", config.Type)
	}
	return nil
}

func initializeTorrentClient(
	ctx context.Context,
	config TorrentConfig,
) torrentclient.Client {
	switch config.Type {
	case TorrentClientTypeQBittorrent:
		config := config.QBittorrent

		return qbittorrent.New(
			ctx,
			config.Host,
			config.Username,
			config.Password,
		)
	default:
		log.Panic().Msgf("torrentClient type %s not implemented", config.Type)
	}
	return nil
}

func initializeTorrentSource(c TorrentSourceConfig) torrentsource.Source {
	httpClient := &http.Client{
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Second), 1),
		),
		Timeout: 15 * time.Second,
	}

	switch c.Type {
	case TorrentSourceTypeNyaa:
		return nyaa.New(httpClient, nyaa.Config{
			CustomParameters: c.Nyaa.CustomParameters,
		})
	case TorrentSourceTypeNekoBT:
		return nekobt.New(httpClient, nekobt.Config{
			APIKey:           c.Nekobt.APIKey,
			CustomParameters: c.Nekobt.CustomParameters,
		})
	default:
		log.Panic().Msgf("rss type %s not implemented", c.Type)
	}
	return nil
}

func initializeShoko(c ShokoConfig) *shokoadapter.API {
	httpClient := &http.Client{
		Transport: defaultTransport,
		Timeout:   15 * time.Second,
	}
	return shokoadapter.New(httpClient, shoko.Config{
		Host:   c.Host,
		APIKey: c.APIKey,
	})
}

func main() {
	log.Info().Msgf("starting Animeman [%s]", version)

	config, err := ReadConfig(
		coalesce.OrDefault(os.Getenv("CONFIG_PATH"), "config.yaml"),
	)
	if err != nil {
		log.Fatal().Msgf("config is not valid: %s", err)
	}

	zerolog.SetGlobalLevel(config.LogLevel.Convert())

	ctx, done := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer done()

	renameScript, err := expr.Compile(config.RenameScript)
	if err != nil {
		log.Fatal().Msgf("failed to compile rename script: %s", err)
	}

	discoveryConfig := config.DiscoveryConfig

	var anidbEpResolver animelist.EpisodeAnidbIDResolver
	var shokoClient shoko.Shoko
	if config.ShokoConfig.Host != "" {
		api := initializeShoko(config.ShokoConfig)
		api.Wait(ctx)

		anidbEpResolver = api
		shokoClient = api
	}

	httpClient := &http.Client{
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Minute), 30),
		),
		Timeout: 15 * time.Second,
	}

	// Anilist API is used for MAL->AniList id resolution,
	// so that shoko can be matched by exact id instead of fuzzy title search.
	anilistAPI := anilist.New(httpClient, config.Username, config.CacheTTL)
	torrentSource := initializeTorrentSource(config.TorrentSourceConfig)

	var anidbResolver animelist.AnidbIDResolver
	if nekobt, ok := torrentSource.(*nekobt.API); ok {
		anidbResolver = nekobt
	}

	deps := controller.Dependencies{
		AnimeListSource:   initializeAnimeList(httpClient, config.AnimeListConfig, anilistAPI),
		TorrentSource:     torrentSource,
		TorrentClient:     initializeTorrentClient(ctx, config.TorrentConfig),
		AnilistIDResolver: anilistAPI,
		AnidbIDResolver:   anidbResolver,
		AnidbEpIDResolver: anidbEpResolver,
		Shoko:             shokoClient,
		Config: controller.Config{
			SearchSuffix:     discoveryConfig.SearchSuffix,
			ReleaseGroups:    discoveryConfig.Sources,
			Qualitites:       discoveryConfig.Qualities,
			Category:         discoveryConfig.Category,
			DownloadPath:     discoveryConfig.DownloadPath,
			CreateShowFolder: discoveryConfig.CreateShowFolder,
			PollFrequency:    discoveryConfig.PollFrequency,
			RenameTorrent:    coalesce.Coalesce(discoveryConfig.RenameTorrent, true),
			RenameFormat:     renameScript,
		},
	}

	controller := controller.New(deps)

	healthServer := newHealthcheck(deps)

	go func() {
		log.Info().Msgf("health server listening on %s", healthServer.Addr)
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Msgf("health server: %s", err)
		}
	}()

	controller.Start()

	<-ctx.Done()

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	errgrp, errctx := errgroup.WithContext(shutdownCtx)

	errgrp.Go(func() error {
		return healthServer.Shutdown(errctx)
	})

	errgrp.Go(func() error {
		return controller.Shutdown(errctx)
	})

	if err := errgrp.Wait(); err != nil {
		log.
			Error().
			Err(err).
			Msg("could not shutdown gracefully")

		return
	}

	log.Info().Msg("shutdown successful")
}
