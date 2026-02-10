package anilist

import (
	"net/http"

	"github.com/sonalys/animeman/internal/domain"
)

const API_URL = "https://graphql.anilist.co"

type (
	API struct {
		Username        string
		client          *http.Client
		cachedAnimeList []domain.Entry
	}
)

func New(client *http.Client, username string) *API {
	return &API{
		client:   client,
		Username: username,
	}
}
