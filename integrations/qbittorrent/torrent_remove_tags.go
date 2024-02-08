package qbittorrent

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

func (api *API) RemoveTorrentTags(ctx context.Context, hashes []string, args ...AddTorrentTagsArg) error {
	var path = api.host + "/torrents/removeTags"
	values := url.Values{
		"hashes": []string{strings.Join(hashes, "|")},
	}
	for _, f := range args {
		f.ApplyAddTorrentTags(values)
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
