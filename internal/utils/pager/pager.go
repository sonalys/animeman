// Package pager implements the paginated search loop shared by the
// torrent source adapters (nyaa, nekobt): fetch pages until they run out,
// filter/map each page into torrents, and stop early once the source has
// no results newer than the latest downloaded tag.
package pager

import (
	"context"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
)

// Page describes one paginated search over an RSS/torznab feed.
type Page[T any] struct {
	// PageSize is the number of items requested per page.
	PageSize int

	// Fetch fetches one page of raw items at the given offset.
	Fetch func(ctx context.Context, offset int) ([]T, error)
	// Filter keeps only the items that should become torrents.
	Filter func(T) bool
	// Map converts a filtered item into a torrent.
	Map func(T) torrentsource.Torrent
	// ShouldPaginate reports whether the source may have more results
	// after this page, given the filtered items of the current page.
	ShouldPaginate func(filtered []T) bool
}

// Search runs the paginated loop and returns the mapped torrents.
func Search[T any](ctx context.Context, p Page[T]) ([]torrentsource.Torrent, error) {
	torrents := make([]torrentsource.Torrent, 0, p.PageSize)
	offset := 0

	for {
		items, err := p.Fetch(ctx, offset)
		if err != nil {
			return nil, err
		}

		if len(items) == 0 {
			break
		}

		offset += len(items)

		filtered := items
		if p.Filter != nil {
			filtered = utils.Filter(items, p.Filter)
		}

		for _, item := range filtered {
			torrents = append(torrents, p.Map(item))
		}

		if len(items) < p.PageSize || !p.ShouldPaginate(filtered) {
			break
		}
	}

	return torrents, nil
}
