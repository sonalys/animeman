// Docs: https://nyaa.si/help
package nyaa

import (
	"net/http"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

const API_URL = "https://nyaa.si/?page=rss"

type (
	Config struct {
		// CustomParameters are extra torznab query parameters, e.g.
		// `sort: seeders`, `sub_lang: en`, `mtl: "false"`. They override
		// the parameters built by the adapter.
		CustomParameters map[string]string
	}

	API struct {
		config Config
		client *http.Client
	}
)

var _ torrentsource.Source = (*API)(nil)

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}
