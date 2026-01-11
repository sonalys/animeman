package myanimelist

import (
	"net/http"

	"github.com/sonalys/animeman/pkg/v1/animelist"
)

const API_URL = "https://myanimelist.net"

type (
	API struct {
		Username        string
		client          *http.Client
		cachedAnimeList []animelist.Entry
	}
)

func New(client *http.Client, username string) *API {
	return &API{
		client:   client,
		Username: username,
	}
}
