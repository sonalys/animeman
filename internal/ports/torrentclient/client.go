package torrentclient

import (
	"context"
	"fmt"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

// TorrentClient is the port implemented by every torrent client adapter (qbittorrent, ...).
type TorrentClient interface {
	List(ctx context.Context, arg *ListTorrentConfig) ([]Torrent, error)
	// TorrentFiles lists the file paths contained in the given torrent hash.
	TorrentFiles(ctx context.Context, hash string) ([]string, error)
	AddTorrent(ctx context.Context, arg *AddTorrentConfig) error
	AddTorrentTags(ctx context.Context, hashes []string, tags []string) error
	RemoveTorrentTags(ctx context.Context, hashes []string, tags []string) error
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
		// Tag filters to torrents with the given tag. Send empty string to filter to torrents with no tags.
		Tag *string
		// Completed filters to fully downloaded torrents only (qbittorrent "completed" filter).
		Completed *bool
	}
)
