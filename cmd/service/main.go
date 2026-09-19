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
	"github.com/sonalys/animeman/cmd/service/configs"
	"github.com/sonalys/animeman/internal/adapters/animelist/anilist"
	"github.com/sonalys/animeman/internal/adapters/animelist/myanimelist"
	shokoadapter "github.com/sonalys/animeman/internal/adapters/shoko"
	"github.com/sonalys/animeman/internal/adapters/torrentclient/qbittorrent"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nekobt"
	"github.com/sonalys/animeman/internal/adapters/torrentsource/nyaa"
	"github.com/sonalys/animeman/internal/discovery"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/shoko"
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

func initializeAnimeList(
	httpClient *http.Client,
	config configs.AnimeListConfig,
	anilistAPI *anilist.API,
) animelist.AnimeListSource {
	switch config.Type {
	case configs.AnimeListTypeMAL:
		return myanimelist.New(httpClient, config.Username, config.CacheTTL)
	case configs.AnimeListTypeAnilist:
		return anilistAPI
	default:
		log.Panic().Msgf("animeListType %s not implemented", config.Type)
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
			APIKey:           c.Nekobt.APIKey,
			CustomParameters: c.Nekobt.CustomParameters,
		})
	default:
		log.Panic().Msgf("rss type %s not implemented", c.Type)
	}
	return nil
}

func initializeShoko(c configs.ShokoConfig) shoko.Shoko {
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

	// Anilist API is used for MAL->AniList id resolution, so that shoko can be matched by exact id instead of fuzzy title search.
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
			RenameTorrent:    utils.Coalesce(discoveryConfig.RenameTorrent, true),
			RenameFormat:     renameScript,
		},
	})
	if err := c.Start(ctx); err != nil {
		log.Error().Msgf("failed to shutdown: %s", err)
	} else {
		log.Info().Msg("shutdown successful")
	}
}
