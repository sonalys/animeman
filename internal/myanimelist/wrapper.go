package myanimelist

import (
	"context"
	"net/http"
	"time"

	"github.com/sonalys/animeman/integrations/myanimelist"
	"github.com/sonalys/animeman/internal/roundtripper"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/pkg/v1/animelist"
	"golang.org/x/time/rate"
)

type (
	Wrapper struct {
		client *myanimelist.API
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
		client: myanimelist.New(client, username),
	}
}

func convertMALEntry(in []myanimelist.AnimeListEntry) []animelist.Entry {
	out := make([]animelist.Entry, 0, len(in))
	for i := range in {
		out = append(out, animelist.Entry{
			ListStatus:   animelist.ListStatus(in[i].Status),
			Title:        in[i].Title,
			TitleEng:     in[i].TitleEng,
			AiringStatus: animelist.AiringStatus(in[i].AiringStatus),
		})
	}
	return out
}

func (w *Wrapper) GetAnimeList(ctx context.Context, args ...animelist.AnimeListArg) ([]animelist.Entry, error) {
	resp, err := w.client.GetAnimeList(ctx, utils.ConvertInterfaceList[animelist.AnimeListArg, myanimelist.AnimeListArg](args)...)
	if err != nil {
		return nil, err
	}
	return convertMALEntry(resp), nil
}
