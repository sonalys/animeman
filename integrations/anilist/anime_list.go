package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/sonalys/animeman/internal/utils"
)

type (
	ListStatus   string
	AiringStatus string
	Title        string

	AnimeListEntry struct {
		Status ListStatus `json:"status"`
		Media  struct {
			Type         string
			AiringStatus AiringStatus `json:"status"`
			Title        struct {
				Romaji  string `json:"romaji"`
				English string `json:"english"`
				Native  string `json:"native"`
			} `json:"title"`
		} `json:"media"`
	}

	GraphqlQuery struct {
		Query     string         `json:"query"`
		Variables map[string]any `json:"variables"`
	}

	AnimeListResp struct {
		Data struct {
			MediaListCollection struct {
				Lists []struct {
					Entries []AnimeListEntry `json:"entries"`
				} `json:"lists"`
			} `json:"MediaListCollection"`
		} `json:"data"`
	}
)

const (
	ListStatusWatching  ListStatus = "CURRENT"
	ListStatusCompleted ListStatus = "COMPLETED"
	ListStatusDropped   ListStatus = "DROPPED"
	ListStatusPlanning  ListStatus = "PLANNING"
)

const (
	AiringStatusAiring    AiringStatus = "AIRING"
	AiringStatusCompleted AiringStatus = "COMPLETED"
)

const getCurrentlyWatchingQuery = `query($userName:String,$type:MediaType){ 
	MediaListCollection(userName:$userName,type:$type){
		lists{
			name
			entries{
				status
				media{
					title{romaji english native}
					type 
					status(version:2)
				}
			}
		}
	}
}
`

func filter[T any](in []T, f func(T) bool) []T {
	out := make([]T, 0, len(in))
	for i := range in {
		if f(in[i]) {
			out = append(out, in[i])
		}
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]AnimeListEntry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	body := GraphqlQuery{
		Query: getCurrentlyWatchingQuery,
		Variables: map[string]any{
			"userName": api.Username,
			"type":     "ANIME",
		},
	}
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(utils.Must(json.Marshal(body)))))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("invalid response: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}
	var entries AnimeListResp
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	var out []AnimeListEntry
	for _, list := range entries.Data.MediaListCollection.Lists {
		watchingEntries := filter(list.Entries, func(entry AnimeListEntry) bool {
			return entry.Status == ListStatusWatching
		})
		out = append(out, watchingEntries...)
	}
	return out, nil
}
