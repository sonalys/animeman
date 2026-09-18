package discovery

import (
	"context"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentclient"
)

type (
	AnimeListSource interface {
		GetCurrentlyWatching(ctx context.Context) ([]animelist.Entry, error)
	}

	TorrentClient interface {
		List(ctx context.Context, arg *torrentclient.ListTorrentConfig) ([]torrentclient.Torrent, error)
		AddTorrent(ctx context.Context, arg *torrentclient.AddTorrentConfig) error
		AddTorrentTags(ctx context.Context, hashes []string, tags []string) error
	}
)
