package main

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/expr-lang/expr"
	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/service/discovery"
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
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

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

func initializeShoko(c ShokoConfig) shoko.Shoko {
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

	var shokoClient shoko.Shoko
	if config.ShokoConfig.Host != "" {
		shokoClient = initializeShoko(config.ShokoConfig)
		shokoClient.Wait(ctx)
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

	c := discovery.New(discovery.Dependencies{
		AnimeListSource:   initializeAnimeList(httpClient, config.AnimeListConfig, anilistAPI),
		TorrentSource:     initializeTorrentSource(config.TorrentSourceConfig),
		TorrentClient:     initializeTorrentClient(ctx, config.TorrentConfig),
		AnilistIDResolver: anilistAPI,
		Shoko:             shokoClient,
		Config: discovery.Config{
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
	})

	healthMux := http.NewServeMux()
	healthMux.HandleFunc("/healthz", func(w http.ResponseWriter, _ *http.Request) {
		checkCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		var failed []string
		deps := c.Deps()
		if _, err := deps.TorrentClient.List(checkCtx, nil); err != nil {
			failed = append(failed, "torrentClient")
		}
		if _, err := deps.AnimeListSource.GetCurrentlyWatching(checkCtx); err != nil {
			failed = append(failed, "animeList")
		}
		if _, err := deps.TorrentSource.Search(
			checkCtx,
			animelist.Entry{},
			torrentsource.SearchOptions{},
		); err != nil {
			failed = append(failed, "torrentSource")
		}

		if len(failed) > 0 {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "unhealthy: %s", strings.Join(failed, ", "))
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	healthMux.HandleFunc("/readyz", func(w http.ResponseWriter, _ *http.Request) {
		if err := c.LastRunErr(); err != nil {
			w.WriteHeader(http.StatusServiceUnavailable)
			fmt.Fprintf(w, "last scan failed: %s", err)
			return
		}
		w.WriteHeader(http.StatusOK)
	})
	healthServer := &http.Server{
		Addr:    coalesce.OrDefault(os.Getenv("HEALTH_ADDR"), ":8080"),
		Handler: healthMux,
	}
	go func() {
		log.Info().Msgf("health server listening on %s", healthServer.Addr)
		if err := healthServer.ListenAndServe(); err != http.ErrServerClosed {
			log.Error().Msgf("health server: %s", err)
		}
	}()
	defer func() {
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		healthServer.Shutdown(shutdownCtx)
	}()

	if err := c.Start(ctx); err != nil {
		log.Error().Msgf("failed to shutdown: %s", err)
	} else {
		log.Info().Msg("shutdown successful")
	}
}
