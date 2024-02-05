package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
)

type ArgListTorrent interface {
	ApplyListTorrent(url.Values)
}

func (api *API) List(ctx context.Context, args ...ArgListTorrent) ([]Torrent, error) {
	var path = api.host + "/torrents/info"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	values := url.Values{}
	for _, f := range args {
		f.ApplyListTorrent(values)
	}
	req.URL.RawQuery = values.Encode()
	resp, err := api.Do(req)
	if err != nil {
		return nil, fmt.Errorf("could not list torrents: %w", err)
	}
	defer resp.Body.Close()
	var respBody []Torrent
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}
	return respBody, nil
}
