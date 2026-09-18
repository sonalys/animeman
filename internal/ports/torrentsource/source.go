package torrentsource

import (
	"context"
	"time"

	"github.com/sonalys/animeman/internal/ports/animelist"
)

// Torrent is a normalized torrent entry returned by any torrent source.
type Torrent struct {
	Title   string
	Link    string
	PubDate time.Time
	Seeders int
}

// SearchOptions controls how a source searches for torrents of an entry.
type SearchOptions struct {
	// SearchSuffix is appended to the search query, e.g. `-"dub"`.
	SearchSuffix string
	// Sources are release groups to filter for, in priority order. Empty accepts all.
	Sources []string
	// Qualities are quality filters, e.g. "1080 HEVC". Empty accepts all.
	Qualities []string
}

// Source is the port implemented by every torrent source (nyaa, nekobt, ...).
// It receives an anime list entry and returns downloadable torrents for it,
// already filtered and sorted so the newest/best candidates come first.
type Source interface {
	Search(ctx context.Context, entry animelist.Entry, opts SearchOptions) ([]Torrent, error)
}
