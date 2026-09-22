// Docs: https://wiki.nekobt.to/technical-details/torznab
package nekobt

import (
	"net/http"
	"sync"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

var (
	TORZNAB_URL = "https://nekobt.to/api/torznab/api"
	// JSON_URL is the base URL of the nekoBT JSON API.
	// https://wiki.nekobt.to/technical-details/json/
	JSON_URL = "https://nekobt.to/api/v1"
)

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

		// mediaIDs caches resolved nekoBT internal media ids, keyed by the
		// external identifier used for the lookup (e.g. `anilist-20594`).
		mediaIDsMu sync.Mutex
		mediaIDs   map[string]string
	}
)

var _ torrentsource.Source = (*API)(nil)

func New(client *http.Client, c Config) *API {
	return &API{
		config:   c,
		client:   client,
		mediaIDs: map[string]string{},
	}
}
