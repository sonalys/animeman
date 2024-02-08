package anilist

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/integrations/anilist"
	"github.com/sonalys/animeman/internal/roundtripper"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"golang.org/x/time/rate"
)

type (
	Wrapper struct {
		client *anilist.API
	}
)

const userAgent = "github.com/sonalys/animeman"

func New(username string) *Wrapper {
	client := &http.Client{
		Transport: roundtripper.NewUserAgentTransport(
			roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(time.Second), 1),
			), userAgent),
		Timeout: 10 * time.Second,
	}
	return &Wrapper{
		client: anilist.New(client, username),
	}
}

func convertStatus(in anilist.ListStatus) animelist.ListStatus {
	switch in {
	case anilist.ListStatusWatching:
		return animelist.ListStatusWatching
	case anilist.ListStatusCompleted:
		return animelist.ListStatusCompleted
	case anilist.ListStatusDropped:
		return animelist.ListStatusDropped
	case anilist.ListStatusPlanning:
		return animelist.ListStatusPlanToWatch
	default:
		log.Fatal().Msgf("unexpected status from anilist: %s", in)
	}
	return animelist.ListStatusAll
}

func convertAiringStatus(in anilist.AiringStatus) animelist.AiringStatus {
	switch in {
	case anilist.AiringStatusAiring:
		return animelist.AiringStatusAiring
	case anilist.AiringStatusCompleted:
		return animelist.AiringStatusAired
	}
	return animelist.AiringStatus(-1)
}

func convertMALEntry(in []anilist.AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	for i := range in {
		titles := in[i].Media.Title
		out = append(out, animelist.Entry{
			ListStatus:   convertStatus(in[i].Status),
			Titles:       []string{titles.English, titles.Romaji, titles.Native},
			AiringStatus: convertAiringStatus(in[i].Media.AiringStatus),
		})
	}
	return out
}

func (w *Wrapper) GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error) {
	resp, err := w.client.GetCurrentlyWatching(ctx)
	if err != nil {
		return nil, err
	}
	return convertMALEntry(resp), nil
}
