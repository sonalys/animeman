package main

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/anilist"
	"github.com/sonalys/animeman/internal/adapters/myanimelist"
	"github.com/sonalys/animeman/internal/app/discovery"
	"github.com/sonalys/animeman/internal/configs"
	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

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
