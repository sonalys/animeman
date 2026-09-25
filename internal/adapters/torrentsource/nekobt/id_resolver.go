package nekobt

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"

	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/ports/animelist"
)

func (api *API) FromAnilist(ctx context.Context, anilistID int) (int, error) {
	externalIDs, err := api.resolveMediaID(ctx, fmt.Sprintf("anilist-%d", anilistID))
	if err != nil {
		return 0, nil
	}

	anidbIDStr, ok := externalIDs["anidb"]
	if !ok {
		return -1, errors.New("not found")
	}

	anidbID, err := strconv.ParseInt(anidbIDStr, 10, 64)
	if err != nil {
		return -1, fmt.Errorf("malformed anidb id: %w", err)
	}

	return int(anidbID), nil
}

// resolveMediaID returns the nekoBT internal media id for the entry, e.g.
// `s123` or `m456`. It resolves the entry's external id (anilist/mal) through
// the JSON API `/media/resolve` endpoint, caching results per external id.
// https://wiki.nekobt.to/technical-details/json/#resolve-external-media-id
func (api *API) resolveMediaID(ctx context.Context, externalID string) (map[string]string, error) {
	api.mediaIDsMu.Lock()
	mediaID, ok := api.mediaIDs[externalID]
	api.mediaIDsMu.Unlock()
	if ok {
		return mediaID, nil
	}

	externalIDs, err := api.fetchMediaID(ctx, externalID)
	if err != nil {
		return nil, err
	}

	if len(externalIDs) == 0 {
		// Unmapped on nekoBT: search by the external id instead.
		source, id, found := strings.Cut(externalID, "-")
		if !found {
			return nil, fmt.Errorf("malformed externalID: %v", externalID)
		}

		return map[string]string{
			source: id,
		}, nil
	}

	api.mediaIDsMu.Lock()
	// Map for each combination of ids the final result.
	for key, value := range externalIDs {
		api.mediaIDs[fmt.Sprintf("%s-%s", key, value)] = externalIDs
	}
	api.mediaIDsMu.Unlock()

	return externalIDs, nil
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
func (api *API) fetchMediaID(ctx context.Context, externalID string) (map[string]string, error) {
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
		return nil, nil
	}
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(body))
	}

	var resolved struct {
		Data map[string]string `json:"data"`
	}

	if err := json.Unmarshal(body, &resolved); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return resolved.Data, nil
}
