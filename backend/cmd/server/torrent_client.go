package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/cmd/server/configs"
	"github.com/sonalys/animeman/internal/adapters/qbittorrent"
	"github.com/sonalys/animeman/internal/app/discovery"
)

func initializeTorrentClient(ctx context.Context, c configs.TorrentConfig) discovery.TorrentClient {
	switch c.Type {
	case configs.TorrentClientTypeQBittorrent:
		client, err := qbittorrent.New(ctx, c.Host, c.Username, c.Password)
		if err != nil {
			log.Fatal().
				Err(err).
				Msg("Could not connect to qBitTorrent")
		}

		return client
	default:
		log.Panic().
			Str("clientType", string(c.Type)).
			Msgf("Torrent client not implemented")
	}
	return nil
}
