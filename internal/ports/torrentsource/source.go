package torrentsource

import (
	"context"

	"github.com/sonalys/animeman/internal/ports/animelist"
)

// Torrent is a normalized torrent entry returned by any torrent source.
type Torrent struct {
	Title   string
	Link    string
	Seeders int
	Hash    string
}

// SearchOptions controls how a source searches for torrents of an entry.
type SearchOptions struct {
	SearchSuffix string
	Sources      []string
	Qualities    []string
}

// Source is the port implemented by every torrent source (nyaa, nekobt, ...).
// It receives an anime list entry and returns downloadable torrents for it,
// already filtered and sorted so the newest/best candidates come first.
type Source interface {
	Search(ctx context.Context, entry animelist.Entry, opts SearchOptions) ([]Torrent, error)
}
