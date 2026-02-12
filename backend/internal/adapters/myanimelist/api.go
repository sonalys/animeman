package myanimelist

import (
	"net/http"

	"github.com/sonalys/animeman/internal/domain"
)

const API_URL = "https://mydomain.net"

type (
	API struct {
		Username        string
		client          *http.Client
		cachedAnimeList []domain.AnimeListEntry
	}
)

func New(client *http.Client, username string) *API {
	return &API{
		client:   client,
		Username: username,
	}
}
