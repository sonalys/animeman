package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
)

// Temporary solution for finding the correct time format for MAL entries.
// The format changes depending on the user location, and no parameter is given from MAL API to which format is being returned.
func findCorrectTimeFormat(in []AnimeListEntry) string {
	masks := []string{"01-02-06", "02-01-06"}
outer:
	for _, mask := range masks {
		for i := range in {
			_, err := time.Parse(mask, in[i].AnimeStartDateString)
			if err != nil {
				continue outer
			}
		}
		return mask
	}
	return masks[0]
}

func convertTitles(in ...string) []string {
	out := make([]string, 0, len(in))
	for i := range in {
		if in[i] == "" {
			continue
		}
		out = append(out, in[i])
	}
	return out
}

func convertEntry(in []AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	timeFormat := findCorrectTimeFormat(in)
	for i := range in {
		out = append(out, animelist.NewEntry(
			convertTitles(fmt.Sprint(in[i].Title), in[i].TitleEng),
			animelist.ListStatus(in[i].Status),
			animelist.AiringStatus(in[i].AiringStatus),
			utils.Must(time.Parse(timeFormat, in[i].AnimeStartDateString)),
			in[i].NumEpisodes,
		))
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	v := url.Values{
		"offset": []string{"0"},
		"status": []string{"1"},
	}
	req.URL.RawQuery = v.Encode()
	resp, err := api.client.Do(req)
	if err != nil {
		if len(api.cachedAnimeList) > 0 {
			log.Warn().Msgf("failed to fetch anime list, using cache: %s", err)
			return api.cachedAnimeList, nil
		}
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	var entries []AnimeListEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	api.cachedAnimeList = convertEntry(entries)
	return api.cachedAnimeList, nil
}
