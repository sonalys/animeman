package shoko

import (
	"context"
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

func (api *API) Wait(ctx context.Context) {
	log.Info().Msgf("probing for shoko")
	for {
		if ctx.Err() != nil {
			return
		}
		req, err := http.NewRequestWithContext(ctx, http.MethodGet, api.config.Host+"/api/v3/Server/Version", nil)
		if err == nil {
			resp, err := api.do(ctx, req)
			if err == nil {
				resp.Body.Close()
				log.Info().Msgf("shoko is ready")
				return
			}
		}
		time.Sleep(time.Second)
	}
}
