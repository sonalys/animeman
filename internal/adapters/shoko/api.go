// Docs: https://shoko.tld/swagger/index.html
package shoko

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"

	"github.com/sonalys/animeman/internal/ports/shoko"
	"github.com/sonalys/animeman/internal/utils"
)

const (
	seriesSearch      = "/api/v3/Series/AniDB/Search"
	seriesByAnilistID = "/api/v3/Anilist/Anime/%d/Shoko/Series"
	episodesPath      = "/api/v3/Series/AniDB/%d/Episode"
	pathEndsWith      = "/api/v3/File/PathEndsWith/%s"
	linkFilePath      = "/api/v3/File/%d/Link"
	rescanPath        = "/api/v3/File/%d/Rescan"
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
	rawBody := utils.Must(io.ReadAll(resp.Body))
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("request failed: %s: %s", resp.Status, string(rawBody))
	}
	if err := json.Unmarshal(rawBody, out); err != nil {
		return fmt.Errorf("reading response: %w", err)
	}
	return nil
}

type (
	anidbAnime struct {
		ID      int    `json:"id"`
		ShokoID *int   `json:"shokoID"`
		Title   string `json:"title"`
	}

	listResult[T any] struct {
		List  []T `json:"list"`
		Total int `json:"total"`
	}

	shokoSeries struct {
		IDs struct {
			AniDB int `json:"aniDB"`
		} `json:"ids"`
	}

	anidbEpisode struct {
		ID            int    `json:"id"`
		EpisodeNumber int    `json:"episodeNumber"`
		Type          string `json:"type"`
	}

	shokoEpisode struct {
		IDs struct {
			ID    int `json:"id"`
			AniDB int `json:"aniDB"`
		} `json:"ids"`
		AniDB anidbEpisode `json:"aniDB"`
	}

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
		// SeriesIDs are the cross-references, only present when include=XRefs.
		// A file linked to at least one episode has non-empty series cross-references.
		SeriesIDs []struct {
			EpisodeIDs []struct {
				ID *shokoXRefIDs `json:"id"`
			} `json:"episodeIDs"`
		} `json:"seriesIDs"`
	}
)

// FindSeriesByAnilistID implements shoko.Shoko.
func (api *API) FindSeriesByAnilistID(ctx context.Context, anilistID int) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+fmt.Sprintf(seriesByAnilistID, anilistID),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := api.do(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return 0, nil
	}

	var series []shokoSeries
	if err := decodeJSON(resp, &series); err != nil {
		return 0, fmt.Errorf("finding series by anilist id: %w", err)
	}
	if len(series) == 0 {
		return 0, nil
	}
	return series[0].IDs.AniDB, nil
}

// FindSeriesByTitle implements shoko.Shoko.
func (api *API) FindSeriesByTitle(ctx context.Context, title string) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+seriesSearch+"/"+url.PathEscape(title),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Set("includeTitles", "true")
	q.Set("pageSize", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}

	var result listResult[anidbAnime]
	if err := decodeJSON(resp, &result); err != nil {
		return 0, fmt.Errorf("searching series: %w", err)
	}
	if len(result.List) == 0 {
		return 0, nil
	}
	return result.List[0].ID, nil
}

// FindEpisodes implements shoko.Shoko.
func (api *API) FindEpisodes(ctx context.Context, anidbID int) ([]shoko.Episode, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+fmt.Sprintf(episodesPath, anidbID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Set("type", "Episode")
	q.Set("pageSize", "0")
	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	var result listResult[shokoEpisode]
	if err := decodeJSON(resp, &result); err != nil {
		return nil, fmt.Errorf("listing episodes: %w", err)
	}

	return utils.Map(result.List, func(in shokoEpisode) shoko.Episode {
		return shoko.Episode{
			AniDBID:        in.AniDB.ID,
			ShokoEpisodeID: in.IDs.ID,
			Number:         in.AniDB.EpisodeNumber,
		}
	}), nil
}

// FindFileByPath implements shoko.Shoko.
// It searches shoko for a file whose path ends with the given suffix,
// returning nil when shoko does not know the file at all.
func (api *API) FindFileByPath(ctx context.Context, pathSuffix string) (*shoko.File, bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+fmt.Sprintf(pathEndsWith, url.PathEscape(pathSuffix)),
		nil,
	)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}
	q := req.URL.Query()
	q.Set("include", "XRefs,ReleaseInfo")
	q.Set("limit", "1")
	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, false, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusNotFound {
		resp.Body.Close()
		return nil, false, nil
	}

	var files []shokoFile
	if err := decodeJSON(resp, &files); err != nil {
		return nil, false, fmt.Errorf("finding file by path: %w", err)
	}
	if len(files) == 0 {
		return nil, false, nil
	}

	in := files[0]
	linked := false
	for _, series := range in.SeriesIDs {
		if len(series.EpisodeIDs) > 0 {
			linked = true
			break
		}
	}

	file := &shoko.File{
		ID:      in.ID,
		Scanned: in.Release != nil,
	}
	if len(in.Locations) > 0 {
		file.RelativePath = in.Locations[0].RelativePath
	}
	return file, linked, nil
}

// RescanFile implements shoko.Shoko.
func (api *API) RescanFile(ctx context.Context, fileID int) error {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(rescanPath, fileID),
		nil,
	)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}

	resp, err := api.do(ctx, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"rescanning file failed: %s: %s",
			resp.Status,
			string(utils.Must(io.ReadAll(resp.Body))),
		)
	}
	return nil
}

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
			string(utils.Must(io.ReadAll(resp.Body))),
		)
	}
}

// LinkFileToEpisodes implements shoko.Shoko.
func (api *API) LinkFileToEpisodes(ctx context.Context, fileID int, episodeIDs []int) error {
	body, err := json.Marshal(map[string]any{"episodeIDs": episodeIDs})
	if err != nil {
		return fmt.Errorf("marshaling body: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(linkFilePath, fileID),
		bytes.NewReader(body),
	)
	if err != nil {
		return fmt.Errorf("creating request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := api.do(ctx, req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf(
			"linking file failed: %s: %s",
			resp.Status,
			string(utils.Must(io.ReadAll(resp.Body))),
		)
	}
	return nil
}
