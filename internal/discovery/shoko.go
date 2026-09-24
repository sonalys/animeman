package discovery

import (
	"context"
	"fmt"
)

// RunShokoIntegration links shoko's unrecognized files to episodes,
// matching them against the anime list entries and their AniDB ids.
// It drains the shoko queue; only torrents that have fully completed in the
// torrent client are processed, since shoko can only see finished files.
// Torrents still downloading are requeued and retried on the next run.
// It runs on its own one-minute schedule (see runShokoLoop).
func (c *Controller) RunShokoIntegration(
	ctx context.Context,
) error {
	files, err := c.dep.Shoko.ListUnknownFiles(ctx)
	if err != nil {
		return fmt.Errorf("listing shoko unmatched files: %w", err)
	}

	for _, file := range files {
		if file.CreatedAt.Before(c.lastShokoUnknownFileTimestamp) {
			continue
		}

		if _, err := c.dep.Shoko.AutoMatchAndSaveFile(ctx, file.ID); err != nil {
			return fmt.Errorf("auto-matching file: %w", err)
		}

		c.lastShokoUnknownFileTimestamp = file.CreatedAt
	}

	return nil
}
