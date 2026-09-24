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
	filePath            = "/api/v3/File"
	autoMatchFilePath   = "/api/v3/ReleaseInfo/File/%d/AutoPreview"
	releaseInfoFilePath = "/api/v3/ReleaseInfo/File/%d"
	releaseProviderPath = "/api/v3/ReleaseInfo/Provider"
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

// AutoMatchFile asks Shoko to run its local filename-based release search.
//
// A 200 response means Shoko found a release and returns the release info.
// A 204 response means no match was found.
func (api *API) AutoMatchFile(
	ctx context.Context,
	fileID int,
	providerIDs []string,
) (*ReleaseInfo, bool, error) {
	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		api.config.Host+fmt.Sprintf(autoMatchFilePath, fileID),
		nil,
	)
	if err != nil {
		return nil, false, fmt.Errorf("creating request: %w", err)
	}

	q := url.Values{
		"providerIDs": providerIDs,
	}
	q.Set("isAutomatic", "true")

	req.URL.RawQuery = q.Encode()

	resp, err := api.do(ctx, req)
	if err != nil {
		return nil, false, fmt.Errorf("request failed: %w", err)
	}

	defer resp.Body.Close()

	switch resp.StatusCode {
	case http.StatusOK:
		var release ReleaseInfo

		if err := json.NewDecoder(resp.Body).Decode(&release); err != nil {
			return nil, false, fmt.Errorf(
				"decoding auto-match response: %w",
				err,
			)
		}

		return &release, true, nil

	case http.StatusNoContent:
		return nil, false, nil

	case http.StatusUnauthorized, http.StatusForbidden:
		return nil, false, shoko.ErrForbidden

	default:
		body := must.Must(io.ReadAll(resp.Body))

		return nil, false, fmt.Errorf(
			"auto matching file failed: %s: %s",
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

// AutoMatchAndSaveFile runs the complete workflow:
//
//  1. Ask Shoko to preview/automatically match the file.
//  2. If a match was found, submit that release information back to Shoko.
//
// It returns matched=false when AutoPreview did not find a release.
func (api *API) AutoMatchAndSaveFile(
	ctx context.Context,
	fileID int,
) (matched bool, err error) {
	providers, err := api.ListReleaseProviders(ctx)
	if err != nil {
		return false, fmt.Errorf("listing release providers: %w", err)
	}

	release, matched, err := api.AutoMatchFile(ctx, fileID, providers)
	if err != nil {
		return false, err
	}

	if !matched {
		return false, nil
	}

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
