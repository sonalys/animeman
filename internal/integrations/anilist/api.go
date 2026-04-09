package anilist

import (
	"net/http"
	"time"

	"github.com/sonalys/animeman/pkg/v1/animelist"
)

const API_URL = "https://graphql.anilist.co"

type (
	API struct {
		Username        string
		client          *http.Client
		cacheTTL        time.Duration
		cachedAnimeList []animelist.Entry
		cachedAt        time.Time
	}
)

func New(client *http.Client, username string, cacheTTL time.Duration) *API {
	return &API{
		client:   client,
		Username: username,
		cacheTTL: cacheTTL,
	}
}
