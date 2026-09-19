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

type torrentFile struct {
	Name     string  `json:"name"`
	Progress float64 `json:"progress"`
}

// TorrentFiles implements torrentclient.TorrentClient.
func (api *API) TorrentFiles(ctx context.Context, hash string) ([]string, error) {
	q := url.Values{}
	q.Set("hash", hash)

	req, err := http.NewRequest(http.MethodGet, api.host+"/torrents/files?"+q.Encode(), nil)
	if err != nil {
		return nil, fmt.Errorf("torrent files request failed: %w", err)
	}
	resp, err := api.Do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("could not list torrent files: %w", err)
	}
	defer resp.Body.Close()

	rawBody := utils.Must(io.ReadAll(resp.Body))
	var respBody []torrentFile
	if err := json.Unmarshal(rawBody, &respBody); err != nil {
		return nil, fmt.Errorf("could not read response: %w", err)
	}

	paths := make([]string, 0, len(respBody))
	for _, file := range respBody {
		// Skip files still downloading.
		if file.Progress < 1 {
			continue
		}
		paths = append(paths, file.Name)
	}
	return paths, nil
}
