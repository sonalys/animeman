package qbittorrent

import (
	"context"
	"time"

	"github.com/rs/zerolog/log"
)

func (api *Client) Wait(ctx context.Context) {
	log.Info().Msgf("probing for qBittorrent")
	for {
		if ctx.Err() != nil {
			return
		}
		if _, err := api.client.Get(api.host + "/app/version"); err == nil {
			log.Info().Msgf("qBittorrent is ready")
			return
		}
		time.Sleep(time.Second)
	}
}
