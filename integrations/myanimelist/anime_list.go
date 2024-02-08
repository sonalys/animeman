package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/utils"
)

type AnimeListArg interface {
	ApplyList(url.Values)
}

func (api *API) GetCurrentlyWatching(ctx context.Context, args ...AnimeListArg) ([]AnimeListEntry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	v := url.Values{
		"offset": []string{"0"},
		"status": []string{"1"},
	}
	for _, arg := range args {
		arg.ApplyList(v)
	}
	req.URL.RawQuery = v.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	var entries []AnimeListEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return entries, nil
}
