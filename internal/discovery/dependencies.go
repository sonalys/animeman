package discovery

import (
	"context"

	"github.com/sonalys/animeman/pkg/v1/animelist"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
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
