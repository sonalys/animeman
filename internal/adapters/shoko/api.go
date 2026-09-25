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
	"time"

	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/ports/shoko"
)

const (
	filePath               = "/api/v3/File"
	releaseInfoPreviewPath = "/api/v3/ReleaseInfo/File/%d/AutoPreview"
	releaseInfoFilePath    = "/api/v3/ReleaseInfo/File/%d"
	releaseProviderPath    = "/api/v3/ReleaseInfo/Provider"
	episodePath            = "/api/v3/Anime/%d/Episodes"
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

		// Release is only present when include=ReleaseInfo.
		// A nil release means Shoko has not run its AniDB hash scan yet.
		Release *json.RawMessage `json:"release"`

		Created time.Time `json:"created"`

		// SeriesIDs are the cross-references, only present when include=XRefs.
		SeriesIDs []struct {
			EpisodeIDs []struct {
				ID *shokoXRefIDs `json:"id"`
			} `json:"episodeIDs"`
		} `json:"seriesIDs"`
	}
)

type (
	ReleaseInfo struct {
		ID           string  `json:"ID"`
		ProviderName string  `json:"ProviderName"`
		ReleaseURI   *string `json:"ReleaseURI"`
		Version      int     `json:"Version"`
		FileSize     int64   `json:"FileSize"`
		Comment      *string `json:"Comment"`

		OriginalFilename string `json:"OriginalFilename"`

		IsCensored   *bool `json:"IsCensored"`
		IsChaptered  bool  `json:"IsChaptered"`
		IsCreditless *bool `json:"IsCreditless"`
		IsCorrupted  bool  `json:"IsCorrupted"`

		Source string `json:"Source"`

		Group ReleaseGroup `json:"Group"`

		Hashes []ReleaseHash `json:"Hashes"`

		MediaInfo any `json:"MediaInfo"`

		CrossReferences []ReleaseCrossReference `json:"CrossReferences"`

		Metadata string `json:"Metadata"`

		Released time.Time `json:"Released"`
		Updated  time.Time `json:"Updated"`
		Created  time.Time `json:"Created"`

		IsPublic      *bool `json:"IsPublic"`
		PreventRescan bool  `json:"PreventRescan"`
		DeferToNext   bool  `json:"DeferToNext"`
	}

	ReleaseGroup struct {
		ID        string `json:"ID"`
		Name      string `json:"Name"`
		ShortName string `json:"ShortName"`
		Source    string `json:"Source"`
	}

	ReleaseHash struct {
		Type  string `json:"Type"`
		Value string `json:"Value"`
	}

	ReleaseCrossReference struct {
		AnidbEpisodeID  int               `json:"AnidbEpisodeID"`
		AnidbAnimeID    int               `json:"AnidbAnimeID"`
		PercentageStart int               `json:"PercentageStart"`
		PercentageEnd   int               `json:"PercentageEnd"`
		ProviderIDs     map[string]string `json:"ProviderIDs"`
	}

	shokoReleaseProvider struct {
		ID        string `json:"ID"`
		Name      string `json:"Name"`
		IsEnabled bool   `json:"IsEnabled"`
	}
)

func (api *API) ListReleaseProviders(ctx context.Context) ([]string, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+releaseProviderPath,
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	if resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden {
		resp.Body.Close()
		return nil, shoko.ErrForbidden
	}

	var providers []shokoReleaseProvider

	if err := decodeJSON(resp, &providers); err != nil {
		return nil, fmt.Errorf("parsing release providers: %w", err)
	}

	providerIDs := make([]string, 0, len(providers))

	for _, provider := range providers {
		if !provider.IsEnabled {
			continue
		}

		providerIDs = append(providerIDs, provider.ID)
	}

	return providerIDs, nil
}

// PreviewReleaseInfo asks Shoko to run its local filename-based release search.
//
// A 200 response means Shoko found a release and returns the release info.
// A 204 response means no match was found.
func (api *API) PreviewReleaseInfo(
	ctx context.Context,
	fileID int,
	providerIDs []string,
) (*ReleaseInfo, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(releaseInfoPreviewPath, fileID),
		nil,
	)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	q := url.Values{
		"providerIDs": providerIDs,
	}
	q.Set("isAutomatic", "true")

	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var release ReleaseInfo

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, fmt.Errorf(
				"decoding release preview response: %w",
				err,
			)
		}

		return &release, nil

	case http.StatusNoContent:
		return nil, nil

	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, shoko.ErrForbidden

	default:
		body := must.Must(io.ReadAll(resp.Body))

		return nil, fmt.Errorf(
			"file release preview failed: %s: %s",
			resp.Status,
			string(body),
		)
	}
}

// SaveReleaseInfo sends the release information back to Shoko.
//
// This is the second step of the AutoPreview -> ReleaseInfo workflow.
func (api *API) SaveReleaseInfo(
	ctx context.Context,
	fileID int,
	release *ReleaseInfo,
) error {
	// The AutoPreview endpoint returns "Offline Importer".
	// The manual/import endpoint expects the user-associated provider name.
	release.ProviderName = "Offline Importer+User"

	body, err := json.Marshal(release)
	if err != nil {
		return fmt.Errorf("encoding release info: %w", err)
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(releaseInfoFilePath, fileID),
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
		rawBody := must.Must(io.ReadAll(resp.Body))

		if resp.StatusCode == http.StatusUnauthorized ||
			resp.StatusCode == http.StatusForbidden {
			return shoko.ErrForbidden
		}

		return fmt.Errorf(
			"saving release info failed: %s: %s",
			resp.Status,
			string(rawBody),
		)
	}

	return nil
}

func (api *API) MatchFileToCrossReference(
	ctx context.Context,
	fileID int,
	anidbID int,
	epID int,
) (matched bool, err error) {
	providers, err := api.ListReleaseProviders(ctx)
	if err != nil {
		return false, fmt.Errorf("listing release providers: %w", err)
	}

	release, err := api.PreviewReleaseInfo(ctx, fileID, providers)
	if err != nil {
		return false, err
	}

	switch {
	case len(release.CrossReferences) > 0:
		filteredReferences := sliceutils.Filter(
			release.CrossReferences,
			func(rcr ReleaseCrossReference) bool {
				return rcr.AnidbAnimeID == anidbID
			},
		)

		if len(filteredReferences) == 1 {
			release.CrossReferences = filteredReferences
		}
	case len(release.CrossReferences) == 0:
		return false, nil
	}

	release.CrossReferences[0].AnidbAnimeID = anidbID
	release.CrossReferences[0].AnidbEpisodeID = epID

	if err := api.SaveReleaseInfo(ctx, fileID, release); err != nil {
		return false, fmt.Errorf(
			"saving auto-matched release for file %d: %w",
			fileID,
			err,
		)
	}

	return true, nil
}

func (api *API) ListUnknownFiles(ctx context.Context) ([]shoko.File, error) {
	values := url.Values{
		"page":         {"1"},
		"pageSize":     {"100"},
		"include_only": {"Unrecognized"},
	}

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.config.Host+filePath, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}

	req.URL.RawQuery = values.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, shoko.ErrForbidden
	default:
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

type shokoEpisode struct {
	ID struct {
		ID    *int `json:"id"`
		AniDB *int `json:"aniDB"`
	} `json:"id"`

	Number int    `json:"number"`
	Type   string `json:"type"`
}

func (api *API) EpisodeID(
	ctx context.Context,
	aniDBAnimeID int,
	episodeNumber int,
) (int, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodGet,
		api.config.Host+fmt.Sprintf(episodePath, aniDBAnimeID),
		nil,
	)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}

	resp, err := api.do(ctx, req)
	if err != nil {
		return 0, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusUnauthorized ||
		resp.StatusCode == http.StatusForbidden {
		return 0, shoko.ErrForbidden
	}

	if resp.StatusCode != http.StatusOK {
		body := must.Must(io.ReadAll(resp.Body))

		return 0, fmt.Errorf(
			"getting episodes failed: %s: %s",
			resp.Status,
			string(body),
		)
	}

	var episodes []shokoEpisode

	if err := json.NewDecoder(resp.Body).Decode(&episodes); err != nil {
		return 0, fmt.Errorf("decoding episodes: %w", err)
	}

	for _, episode := range episodes {
		if episode.Number != episodeNumber {
			continue
		}

		if episode.ID.AniDB == nil {
			continue
		}

		return *episode.ID.AniDB, nil
	}

	return 0, fmt.Errorf(
		"AniDB episode %d not found for anime %d",
		episodeNumber,
		aniDBAnimeID,
	)
}
