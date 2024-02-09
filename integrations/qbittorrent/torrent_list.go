package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

func convertTorrent(in []Torrent) []torrentclient.Torrent {
	out := make([]torrentclient.Torrent, 0, len(in))
	for i := range in {
		out = append(out, torrentclient.Torrent{
			Name:     in[i].Name,
			Category: in[i].Category,
			Hash:     in[i].Hash,
			Tags:     in[i].GetTags(),
		})
	}
	return out
}

func digestListTorrentArg(arg *torrentclient.ListTorrentConfig) url.Values {
	v := url.Values{}
	if arg.Category != "" {
		v.Set("category", arg.Category)
	}
	if arg.Tag != "" {
		v.Set("tag", arg.Tag)
	}
	return v
}

func (api *API) List(ctx context.Context, arg *torrentclient.ListTorrentConfig) ([]torrentclient.Torrent, error) {
	var path = api.host + "/torrents/info"
	req, err := http.NewRequest(http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("list request failed: %w", err)
	}
	req.URL.RawQuery = digestListTorrentArg(arg).Encode()
	resp, err := api.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not list torrents: %w", err)
	}
	defer resp.Body.Close()
	rawBody := utils.Must(io.ReadAll(resp.Body))
	var respBody []Torrent
	if err := json.Unmarshal(rawBody, &respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %s: %w", string(rawBody), err)
	}
	return convertTorrent(respBody), nil
}
