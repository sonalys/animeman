package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/qbittorrent"
	"github.com/sonalys/animeman/internal/app/discovery"
	"github.com/sonalys/animeman/internal/configs"
)

func initializeTorrentClient(ctx context.Context, c configs.TorrentConfig) discovery.TorrentClient {
	switch c.Type {
	case configs.TorrentClientTypeQBittorrent:
		return qbittorrent.New(ctx, c.Host, c.Username, c.Password)
	default:
		log.Panic().Msgf("animeListType %s not implemented", c.Type)
	}
	return nil
}
