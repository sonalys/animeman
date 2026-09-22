package torrentsource

import (
	"context"

	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/tags"
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
	// LatestTag is the newest season/episode tag already downloaded in the
	// torrent client. Sources may paginate past it when the first page only
	// contains older episodes.
	LatestTag tags.Tag
}

// Source is the port implemented by every torrent source (nyaa, nekobt, ...).
// It receives an anime list entry and returns downloadable torrents for it,
// already filtered and sorted so the newest/best candidates come first.
type Source interface {
	Search(ctx context.Context, entry animelist.Entry, opts SearchOptions) ([]Torrent, error)
}
