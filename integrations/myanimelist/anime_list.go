package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

func convertEntry(in []AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	for i := range in {
		out = append(out, animelist.Entry{
			ListStatus:   animelist.ListStatus(in[i].Status),
			Titles:       []string{fmt.Sprint(in[i].Title), in[i].TitleEng},
			AiringStatus: animelist.AiringStatus(in[i].AiringStatus),
		})
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	v := url.Values{
		"offset": []string{"0"},
		"status": []string{"1"},
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
	return convertEntry(entries), nil
}
