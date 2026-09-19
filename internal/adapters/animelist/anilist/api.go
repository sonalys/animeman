// Docs: https://docs.anilist.co/guide/graphql
package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"sync"
	"time"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/utils"
)

const API_URL = "https://graphql.anilist.co"

type (
	API struct {
		Username        string
		client          *http.Client
		cacheTTL        time.Duration
		cachedAnimeList []animelist.Entry
		cachedAt        time.Time
		malIDLookup     map[int]int
		lock            sync.Mutex
	}
)

func New(client *http.Client, username string, cacheTTL time.Duration) *API {
	return &API{
		client:      client,
		Username:    username,
		cacheTTL:    cacheTTL,
		malIDLookup: make(map[int]int),
	}
}

// GetAnilistIDByMALID implements animelist.AnilistIDResolver.
func (api *API) GetAnilistIDByMALID(ctx context.Context, malID int) (int, error) {
	api.lock.Lock()
	defer api.lock.Unlock()

	if id, ok := api.malIDLookup[malID]; ok {
		return id, nil
	}

	reqBody := GraphqlQuery{
		Query: `query($idMal:Int){ Media(idMal:$idMal, type:ANIME){ id } }`,
		Variables: map[string]any{
			"idMal": malID,
		},
	}

	req, err := http.NewRequestWithContext(
		ctx,
		http.MethodPost,
		API_URL,
		bytes.NewReader(utils.Must(json.Marshal(reqBody))),
	)
	if err != nil {
		return 0, fmt.Errorf("creating request: %w", err)
	}
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return 0, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return 0, fmt.Errorf("invalid response: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var respBody struct {
		Data struct {
			Media struct {
				ID int `json:"id"`
			} `json:"Media"`
		} `json:"data"`
	}
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return 0, fmt.Errorf("reading response: %w", err)
	}

	api.malIDLookup[malID] = respBody.Data.Media.ID
	return respBody.Data.Media.ID, nil
}
