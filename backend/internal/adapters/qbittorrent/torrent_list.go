package qbittorrent

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/errutils"
)

func convertTorrent(in []Torrent) []domain.Torrent {
	out := make([]domain.Torrent, 0, len(in))
	for i := range in {
		out = append(out, domain.Torrent{
			Name:     in[i].Name,
			Category: in[i].Category,
			Hash:     in[i].Hash,
			Tags:     in[i].GetTags(),
		})
	}
	return out
}

func digestListTorrentArg(arg *domain.ListTorrentConfig) url.Values {
	v := url.Values{}
	if arg.Category != nil {
		v.Set("category", *arg.Category)
	}
	if arg.Tag != nil {
		v.Set("tag", *arg.Tag)
	}
	return v
}

func (api *Client) List(ctx context.Context, arg *domain.ListTorrentConfig) ([]domain.Torrent, error) {
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
	rawBody := errutils.Must(io.ReadAll(resp.Body))
	var respBody []Torrent
	if err := json.Unmarshal(rawBody, &respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %s: %w", string(rawBody), err)
	}
	return convertTorrent(respBody), nil
}
