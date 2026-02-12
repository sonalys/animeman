package discovery

import (
	"context"

	"github.com/sonalys/animeman/internal/domain"
)

type (
	AnimeListSource interface {
		GetCurrentlyWatching(ctx context.Context) ([]domain.AnimeListEntry, error)
	}

	TorrentClient interface {
		List(ctx context.Context, arg *domain.ListTorrentConfig) ([]domain.Torrent, error)
		AddTorrent(ctx context.Context, arg *domain.AddTorrentConfig) error
		AddTorrentTags(ctx context.Context, hashes []string, tags []string) error
	}
)
