package nekobt

import (
	"net/http"
)

const TORZNAB_URL = "https://nekobt.to/api/torznab/api"

type Config struct {
	// APIKey is the nekoBT api key, used as the `apikey` query parameter.
	APIKey string
}

type API struct {
	config Config
	client *http.Client
}

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}
