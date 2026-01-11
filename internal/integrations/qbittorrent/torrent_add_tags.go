package qbittorrent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (api *API) AddTorrentTags(ctx context.Context, hashes []string, tags []string) error {
	var path = api.host + "/torrents/addTags"
	values := url.Values{
		"hashes": []string{strings.Join(hashes, "|")},
		"tags":   []string{strings.Join(tags, ",")},
	}
	req, err := http.NewRequest(http.MethodPost, path, strings.NewReader(values.Encode()))
	if err != nil {
		return fmt.Errorf("list request failed: %w", err)
	}
	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")
	resp, err := api.Do(ctx, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	resp.Body.Close()
	return nil
}
