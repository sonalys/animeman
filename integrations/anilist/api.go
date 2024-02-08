package anilist

import "net/http"

const API_URL = "https://graphql.anilist.co"

type (
	API struct {
		Username string
		client   *http.Client
	}
)

func New(client *http.Client, username string) *API {
	return &API{
		client:   client,
		Username: username,
	}
}
