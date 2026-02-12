package myanimelist

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/domain"
	"github.com/sonalys/animeman/internal/utils/errutils"
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

func convertEntry(in []AnimeListEntry) []domain.AnimeListEntry {
	out := make([]domain.AnimeListEntry, 0, len(in))
	timeFormat := findCorrectTimeFormat(in)
	for i := range in {
		out = append(out, domain.NewEntry(
			convertTitles(fmt.Sprint(in[i].Title), in[i].TitleEng),
			domain.ListStatus(in[i].Status),
			domain.AiringStatus(in[i].AiringStatus),
			errutils.Must(time.Parse(timeFormat, in[i].AnimeStartDateString)),
			in[i].NumEpisodes,
		))
	}
	return out
}

func (api *API) GetCurrentlyWatching(ctx context.Context) ([]domain.AnimeListEntry, error) {
	var path = API_URL + "/animelist/" + api.Username + "/load.json"

	req := errutils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))
	v := url.Values{
		"offset": []string{"0"},
		"status": []string{"1"},
	}
	req.URL.RawQuery = v.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode >= 300 {
		if len(api.cachedAnimeList) > 0 {
			log.
				Warn().
				Err(err).
				Msg("mydomain.net api errored, using cached response")
			return api.cachedAnimeList, nil
		}

		return nil, fmt.Errorf("invalid response: %s", string(errutils.Must(io.ReadAll(resp.Body))))
	}

	var entries []AnimeListEntry
	if err := json.NewDecoder(resp.Body).Decode(&entries); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	api.cachedAnimeList = convertEntry(entries)
	return api.cachedAnimeList, nil
}
