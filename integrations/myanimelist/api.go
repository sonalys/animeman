package myanimelist

import (
	"net/http"
)

const API_URL = "https://myanimelist.net"

type (
	API struct {
		Username string
		client   *http.Client
	}
)

func New(client *http.Client, username string) *API {
	return &API{
		Username: username,
		client:   client,
	}
}
