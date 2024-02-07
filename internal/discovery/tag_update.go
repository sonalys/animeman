package discovery

import (
	"context"
	"fmt"

	"github.com/sonalys/animeman/integrations/qbittorrent"
	"github.com/sonalys/animeman/internal/parser"
)

func (c *Controller) UpdateExistingTorrentsTags(ctx context.Context) error {
	torrents, err := c.dep.QB.List(ctx, qbittorrent.Category(c.dep.Config.Category))
	if err != nil {
		return fmt.Errorf("listing: %w", err)
	}
	for _, torrent := range torrents {
		parsedTitle := parser.ParseTitle(torrent.Name)
		if err := c.dep.QB.AddTorrentTags(ctx, []string{torrent.Hash}, parsedTitle.BuildTorrentTags()); err != nil {
			return fmt.Errorf("updating tags: %w", err)
		}
	}
	return nil
}
