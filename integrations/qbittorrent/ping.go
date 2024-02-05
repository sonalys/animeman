package qbittorrent

import (
	"time"

	"github.com/rs/zerolog/log"
)

func (api *API) Wait() {
	for retries := 5; retries >= 0; retries-- {
		_, err := api.client.Get(api.host + "/app/version")
		if err == nil {
			return
		}
		log.Info().Msgf("waiting for qBitTorrent")
		time.Sleep(time.Duration(6-retries) * 3 * time.Second)
	}
}
