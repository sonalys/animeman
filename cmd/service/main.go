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
	"github.com/sonalys/animeman/internal/adapters/animelist/anilist"
	"github.com/sonalys/animeman/internal/adapters/animelist/myanimelist"
	"github.com/sonalys/animeman/internal/adapters/torrentclient/qbittorrent"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nekobt"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nyaa"
	"github.com/sonalys/animeman/internal/configs"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
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
		Out:        os.Stderr,
		TimeFormat: time.RFC3339,
	})

	zerolog.DefaultContextLogger = &log.Logger
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}

func initializeAnimeList(c configs.AnimeListConfig) animelist.AnimeListSource {
	httpClient := &http.Client{
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Minute), 30),
		),
		Timeout: 15 * time.Second,
	}

	switch c.Type {
	case configs.AnimeListTypeMAL:
		return myanimelist.New(httpClient, c.Username, c.CacheTTL)
	case configs.AnimeListTypeAnilist:
		return anilist.New(httpClient, c.Username, c.CacheTTL)
	default:
		log.Panic().Msgf("animeListType %s not implemented", c.Type)
	}
	return nil
}

func initializeTorrentClient(
	ctx context.Context,
	config configs.TorrentConfig,
) torrentclient.TorrentClient {
	switch config.Type {
	case configs.TorrentClientTypeQBittorrent:
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

func initializeTorrentSource(c configs.TorrentSourceConfig) torrentsource.Source {
	httpClient := &http.Client{
		Transport: roundtripper.NewRateLimitedTransport(
			defaultTransport,
			rate.NewLimiter(rate.Every(time.Second), 1),
		),
		Timeout: 15 * time.Second,
	}

	switch c.Type {
	case configs.TorrentSourceTypeNyaa:
		return nyaa.New(httpClient, nyaa.Config{
			ListParameters: c.Nyaa.CustomParameters,
		})
	case configs.TorrentSourceTypeNekoBT:
		return nekobt.New(httpClient, nekobt.Config{
			APIKey: c.Nekobt.APIKey,
		})
	default:
		log.Panic().Msgf("rss type %s not implemented", c.Type)
	}
	return nil
}

func main() {
	log.Info().Msgf("starting Animeman [%s]", version)

	config, err := configs.ReadConfig(
		utils.ValueOrDefault(os.Getenv("CONFIG_PATH"), "config.yaml"),
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

	c := discovery.New(discovery.Dependencies{
		Source:          initializeTorrentSource(config.TorrentSourceConfig),
		AnimeListClient: initializeAnimeList(config.AnimeListConfig),
		TorrentClient:   initializeTorrentClient(ctx, config.TorrentConfig),
		Config: discovery.Config{
			SearchSuffix:     discoveryConfig.SearchSuffix,
			ReleaseGroups:    discoveryConfig.Sources,
			Qualitites:       discoveryConfig.Qualities,
			Category:         discoveryConfig.Category,
			DownloadPath:     discoveryConfig.DownloadPath,
			CreateShowFolder: discoveryConfig.CreateShowFolder,
			PollFrequency:    discoveryConfig.PollFrequency,
			RenameTorrent:    utils.PointerOrDefault(discoveryConfig.RenameTorrent, true),
			RenameFormat:     renameScript,
		},
	})
	if err := c.Start(ctx); err != nil {
		log.Error().Msgf("failed to shutdown: %s", err)
	} else {
		log.Info().Msg("shutdown successful")
	}
}
