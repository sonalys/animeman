package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/utils"
)

func (api *API) GetAnimeList(ctx context.Context, status ListStatus) ([]AnimeListEntry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	req.URL.RawQuery = url.Values{
		"offset": []string{"0"},
		"status": []string{fmt.Sprint(status)},
	}.Encode()
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
