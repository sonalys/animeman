package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://myanimelist.net"

type (
	API struct {
		Username string
		client   *http.Client
	}

	AnimeListEntry struct {
		Status   ListStatus `json:"status"`
		Title    any        `json:"anime_title"`
		TitleEng string     `json:"anime_title_eng"`
	}

	ListStatus int
	Title      string
)

const (
	ListStatusWatching    ListStatus = 1
	ListStatusCompleted   ListStatus = 2
	ListStatusOnHold      ListStatus = 3
	ListStatusDropped     ListStatus = 4
	ListStatusPlanToWatch ListStatus = 6
	ListStatusAll         ListStatus = 7
)

var statusNames = []string{
	"Unknown",
	"Watching",
	"Completed",
	"OnHold",
	"Dropped",
	"WontWatch",
	"PlanToWatch",
	"All",
}

func (s ListStatus) Name() string {
	return statusNames[s]
}

func (e *AnimeListEntry) GetTitle() string {
	if e.TitleEng != "" {
		return e.TitleEng
	}
	return fmt.Sprint(e.Title)
}

func New(username string) *API {
	return &API{
		Username: username,
		client: &http.Client{
			Transport: roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(60*time.Second), 60),
			),
			Timeout: 3 * time.Second,
		},
	}
}

func (api *API) GetAnimeList(ctx context.Context, status ListStatus) ([]AnimeListEntry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	req.URL.RawQuery = url.Values{
		"offset": []string{"0"},
		"status": []string{fmt.Sprint(status)},
	}.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	var entries []AnimeListEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return entries, nil
}
