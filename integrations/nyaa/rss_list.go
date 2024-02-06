package nyaa

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"

	"github.com/sonalys/animeman/internal/utils"
)

func (api *API) List(ctx context.Context, args ...ListArg) ([]Entry, error) {
	var path = API_URL
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	for _, f := range args {
		f.Apply(req)
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}
	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return feed.Channel.Entries, nil
}
