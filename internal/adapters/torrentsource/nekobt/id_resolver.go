package nekobt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/ports/animelist"
)

type IDMap struct {
	MediaID   string `json:"media_id"`
	AnilistID int    `json:"anilist_id"`
	AnidbID   int    `json:"anidb_id"`
}

func (api *API) FromAnilist(ctx context.Context, anilistID int) (int, error) {
	externalIDs, err := api.resolveMediaID(ctx, fmt.Sprintf("anilist-%d", anilistID))
	if err != nil {
		return 0, nil
	}
	return externalIDs.AnidbID, nil
}

// resolveMediaID returns the nekoBT internal media id for the entry, e.g.
// `s123` or `m456`. It resolves the entry's external id (anilist/mal) through
// the JSON API `/media/resolve` endpoint, caching results per external id.
// https://wiki.nekobt.to/technical-details/json/#resolve-external-media-id
func (api *API) resolveMediaID(ctx context.Context, externalID string) (*IDMap, error) {
	api.mediaIDsMu.Lock()
	defer api.mediaIDsMu.Unlock()

	idMap, ok := api.idMapCache[externalID]
	if ok {
		return idMap, nil
	}

	resolvedIDMap, err := api.fetchMediaID(ctx, externalID)
	if err != nil {
		return nil, fmt.Errorf("fetching external ids: %w", err)
	}

	api.idMapCache[externalID] = resolvedIDMap

	return resolvedIDMap, nil
}

// externalMediaID returns the nekoBT external identifier for the entry,
// e.g. `anilist-20594`.
func externalMediaID(entry animelist.Entry) (string, error) {
	switch {
	case entry.AnilistID != 0:
		return fmt.Sprintf("anilist-%d", entry.AnilistID), nil
	case entry.MALID != 0:
		return fmt.Sprintf("mal-%d", entry.MALID), nil
	default:
		return "", fmt.Errorf("entry %q has no anilist or mal id", strings.Join(entry.Titles, ", "))
	}
}

// fetchMediaID resolves an external identifier to the nekoBT internal media
// id via the JSON API. The response is either a plain string id or an object
// with an `id` field, depending on the resolved media type.
func (api *API) fetchMediaID(ctx context.Context, externalID string) (*IDMap, error) {
	req := must.Must(
		http.NewRequestWithContext(ctx, http.MethodGet, JSON_URL+"/media/resolve", nil),
	)
	q := req.URL.Query()
	q.Set("id", externalID)
	req.URL.RawQuery = q.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	body := must.Must(io.ReadAll(resp.Body))
	if resp.StatusCode == 404 {
		// Not mapped on nekoBT: cache the miss so we don't re-query every
		// scan, and fall back to the external id for the torznab search.
		return nil, errors.New("not found")
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(body))
	}

	var resolved struct {
		Data IDMap `json:"data"`
	}

	if err := json.Unmarshal(body, &resolved); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return &resolved.Data, nil
}
