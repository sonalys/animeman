// Docs: https://shoko.tld/swagger/index.html
package shoko

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/ports/shoko"
)

const (
	seriesSearch      = "/api/v3/Series/AniDB/Search"
	seriesByAnilistID = "/api/v3/Anilist/Anime/%d/Shoko/Series"
	episodesPath      = "/api/v3/Series/AniDB/%d/Episode"
	pathEndsWith      = "/api/v3/File/PathEndsWith/%s"
	linkFilePath      = "/api/v3/File/%d/Link"
	rescanPath        = "/api/v3/File/%d/Rescan"
	filePath          = "/api/v3/File"
	autoMatchFilePath = "/api/v3/ReleaseInfo/File/%d/AutoPreview"
)

type (
	API struct {
		config shoko.Config
		client *http.Client
	}
)

var _ shoko.Shoko = (*API)(nil)

func New(client *http.Client, config shoko.Config) *API {
	return &API{
		config: config,
		client: client,
	}
}

// do performs an authenticated request using the configured api key.
func (api *API) do(ctx context.Context, req *http.Request) (*http.Response, error) {
	localReq := req.Clone(ctx)
	localReq.Header.Set("apikey", api.config.APIKey)
	return api.client.Do(localReq)
}

func decodeJSON(resp *http.Response, out any) error {
	defer resp.Body.Close()
	rawBody := must.Must(io.ReadAll(resp.Body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s: %s", resp.Status, string(rawBody))
	}
	if err := json.Unmarshal(rawBody, out); err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	return nil
}

type (
	shokoXRefIDs struct {
		ID    *int `json:"id"`
		AniDB *int `json:"aniDB"`
	}

	shokoFile struct {
		ID        int `json:"id"`
		Locations []struct {
			RelativePath string `json:"relativePath"`
		} `json:"locations"`
		// Release is only present when include=ReleaseInfo. A nil release means
		// shoko has not run its AniDB hash scan on the file yet.
		Release *json.RawMessage `json:"release"`
		Created time.Time        `json:"created"`
		// SeriesIDs are the cross-references, only present when include=XRefs.
		// A file linked to at least one episode has non-empty series cross-references.
		SeriesIDs []struct {
			EpisodeIDs []struct {
				ID *shokoXRefIDs `json:"id"`
			} `json:"episodeIDs"`
		} `json:"seriesIDs"`
	}
)

// AutoMatchFile implements shoko.Shoko.
// It asks shoko to run its local filename-based release search on the file.
// A 200 response means shoko found a release, 204 means no match.
func (api *API) AutoMatchFile(ctx context.Context, fileID int) (bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(autoMatchFilePath, fileID),
		nil,
	)
	if err != nil {
		return false, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Set("isAutomatic", "true")
	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return false, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		return true, nil
	case http.StatusNoContent:
		return false, nil
	default:
		return false, fmt.Errorf(
			"auto matching file failed: %s: %s",
			resp.Status,
			string(must.Must(io.ReadAll(resp.Body))),
		)
	}
}

// ListUnknownFiles implements [shoko.Shoko].
func (api *API) ListUnknownFiles(ctx context.Context) ([]shoko.File, error) {
	values := url.Values{
		"include_only": {"Unrecognized", "ImportLimbo"},
		"sort_order":   {"FileName"},
		"pageSize":     {"200"},
		"page":         {"1"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.config.Host+filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.Header.Add("Accept", "application/json")

	req.URL.RawQuery = values.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf(
			"linking file failed: %s: %s",
			resp.Status,
			string(must.Must(io.ReadAll(resp.Body))),
		)
	}

	var body struct {
		Total int         `json:"total"`
		List  []shokoFile `json:"list"`
	}

	if err := decodeJSON(resp, &body); err != nil {
		return nil, fmt.Errorf("parsing resp: %w", err)
	}

	files := sliceutils.Map(body.List, func(item shokoFile) shoko.File {
		return shoko.File{
			ID:           item.ID,
			RelativePath: item.Locations[0].RelativePath,
			CreatedAt:    item.Created,
		}
	})

	return files, nil
}
