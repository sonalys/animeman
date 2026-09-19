package torrentclient

import (
	"context"
	"fmt"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

// TorrentClient is the port implemented by every torrent client adapter (qbittorrent, ...).
type TorrentClient interface {
	List(ctx context.Context, arg *ListTorrentConfig) ([]Torrent, error)
	AddTorrent(ctx context.Context, arg *AddTorrentConfig) error
	AddTorrentTags(ctx context.Context, hashes []string, tags []string) error
}

type (
	Torrent struct {
		Name     string
		Category string
		Hash     string
		Tags     []string
	}

	AddTorrentConfig struct {
		URLs     []string
		Tags     []string
		Name     *string
		SavePath string
		Category string
		Paused   bool
	}

	ListTorrentConfig struct {
		Category *string
		Tag      *string
	}
)
