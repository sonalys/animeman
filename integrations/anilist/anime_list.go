package anilist

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
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
			Episodes     int          `json:"episodes"`
			StartDate    struct {
				Year  int `json:"year"`
				Month int `json:"month"`
				Day   int `json:"day"`
			} `json:"startDate"`
			Title struct {
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
					startDate{
						year
						month
						day
					}
					title{romaji english native}
					type 
					status(version:2)
					episodes
				}
			}
		}
	}
}
`

func convertStatus(in ListStatus) animelist.ListStatus {
	switch in {
	case ListStatusWatching:
		return animelist.ListStatusWatching
	case ListStatusCompleted:
		return animelist.ListStatusCompleted
	case ListStatusDropped:
		return animelist.ListStatusDropped
	case ListStatusPlanning:
		return animelist.ListStatusPlanToWatch
	default:
		log.Fatal().Msgf("unexpected status from anilist: %s", in)
	}
	return animelist.ListStatusAll
}

func convertAiringStatus(in AiringStatus) animelist.AiringStatus {
	switch in {
	case AiringStatusAiring:
		return animelist.AiringStatusAiring
	case AiringStatusCompleted:
		return animelist.AiringStatusAired
	}
	return animelist.AiringStatus(-1)
}

func convertEntry(in []AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	for i := range in {
		titles := in[i].Media.Title
		out = append(out, animelist.Entry{
			ListStatus:   convertStatus(in[i].Status),
			Titles:       []string{titles.English, titles.Romaji, titles.Native},
			AiringStatus: convertAiringStatus(in[i].Media.AiringStatus),
			StartDate:    time.Date(in[i].Media.StartDate.Year, time.Month(in[i].Media.StartDate.Month), in[i].Media.StartDate.Day, 0, 0, 0, 0, time.UTC),
			NumEpisodes:  in[i].Media.Episodes,
		})
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error) {
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
		watchingEntries := utils.Filter(list.Entries, func(entry AnimeListEntry) bool {
			return entry.Status == ListStatusWatching
		})
		out = append(out, watchingEntries...)
	}
	return convertEntry(out), nil
}
