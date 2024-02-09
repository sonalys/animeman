package discovery

import (
	"context"
	"fmt"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/pkg/v1/torrentclient"
)

// UpdateExistingTorrentsTags will scan all torrents from the configured category and update their tags.
// This function exists for when you already have a collection of Anime categorized torrents.
// This function will tag all entries from the configured category for smart episode detection and filtering.
func (c *Controller) UpdateExistingTorrentsTags(ctx context.Context) error {
	torrents, err := c.dep.TorrentClient.List(ctx, &torrentclient.ListTorrentConfig{
		Category: c.dep.Config.Category,
	})
	if err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for _, torrent := range torrents {
		parsedTitle := parser.TitleParse(torrent.Name)
		if err := c.dep.TorrentClient.AddTorrentTags(ctx, []string{torrent.Hash}, parsedTitle.TagsBuildTorrent()); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}
	return nil
}
