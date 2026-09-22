// Docs: https://wiki.nekobt.to/technical-details/torznab
package nekobt

import (
	"net/http"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

var TORZNAB_URL = "https://nekobt.to/api/torznab/api"

type (
	Config struct {
		// APIKey is the nekoBT api key, used as the `apikey` query parameter.
		APIKey string
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
