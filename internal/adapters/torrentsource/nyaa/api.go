package nyaa

import (
	"net/http"
)

const API_URL = "https://nyaa.si/?page=rss"

type (
	Config struct {
		ListParameters map[string]string
	}

	API struct {
		config Config
		client *http.Client
	}
)

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}
