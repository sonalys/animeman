package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/utils"
)

type ArgListTorrent interface {
	ApplyListTorrent(url.Values)
}

func (t Tag) ApplyListTorrent(v url.Values) {
	v.Add("tag", string(t))
}

func (c Category) ApplyListTorrent(v url.Values) {
	v.Add("category", string(c))
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
	rawBody := utils.Must(io.ReadAll(resp.Body))
	var respBody []Torrent
	if err := json.Unmarshal(rawBody, &respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %s: %w", string(rawBody), err)
	}
	return respBody, nil
}
