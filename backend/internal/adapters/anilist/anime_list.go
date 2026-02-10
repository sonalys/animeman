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
	"github.com/sonalys/animeman/internal/domain"
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
	AiringStatusFinished  AiringStatus = "FINISHED"
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

func convertStatus(in ListStatus) domain.ListStatus {
	switch in {
	case ListStatusWatching:
		return domain.ListStatusWatching
	case ListStatusCompleted:
		return domain.ListStatusCompleted
	case ListStatusDropped:
		return domain.ListStatusDropped
	case ListStatusPlanning:
		return domain.ListStatusPlanToWatch
	default:
		log.Fatal().Msgf("unexpected status from anilist: %s", in)
	}
	return domain.ListStatusAll
}

func convertAiringStatus(in AiringStatus) domain.AiringStatus {
	switch in {
	case AiringStatusAiring:
		return domain.AiringStatusAiring
	case AiringStatusCompleted, AiringStatusFinished:
		return domain.AiringStatusAired
	}
	return domain.AiringStatus(-1)
}

func convertEntry(in []AnimeListEntry) []domain.Entry {
	out := make([]domain.Entry, 0, len(in))
	for i := range in {
		titles := in[i].Media.Title

		out = append(out, domain.NewEntry(
			[]string{titles.English, titles.Romaji, titles.Native},
			convertStatus(in[i].Status),
			convertAiringStatus(in[i].Media.AiringStatus),
			time.Date(in[i].Media.StartDate.Year, time.Month(in[i].Media.StartDate.Month), in[i].Media.StartDate.Day, 0, 0, 0, 0, time.UTC),
			in[i].Media.Episodes,
		))
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]domain.Entry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"

	reqBody := GraphqlQuery{
		Query: getCurrentlyWatchingQuery,
		Variables: map[string]any{
			"userName": api.Username,
			"type":     "ANIME",
		},
	}

	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodPost, path, bytes.NewReader(utils.Must(json.Marshal(reqBody)))))
	req.Header.Add("Content-Type", "application/json")
	req.Header.Add("Accept", "application/json")

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		if len(api.cachedAnimeList) > 0 {
			log.
				Warn().
				Err(err).
				Msg("anilist.co api errored, using cached response")
			return api.cachedAnimeList, nil
		}

		return nil, fmt.Errorf("invalid response: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var respBody AnimeListResp
	if err := json.NewDecoder(resp.Body).Decode(&respBody); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	out := make([]AnimeListEntry, 0, len(respBody.Data.MediaListCollection.Lists))

	for _, list := range respBody.Data.MediaListCollection.Lists {
		watchingEntries := utils.Filter(list.Entries, func(entry AnimeListEntry) bool {
			return entry.Status == ListStatusWatching
		})
		out = append(out, watchingEntries...)
	}

	response := convertEntry(out)
	api.cachedAnimeList = response

	return response, nil
}
