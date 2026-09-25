package controller

import (
	"context"
	"errors"
	"fmt"
	"path"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/pkg/metadata"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/ports/shoko"
)

func (c *Controller) startShokoRoutine(ctx context.Context, shutdown <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastShokoUnknownFileTimestamp time.Time

	for {
		switch nextTimestamp, err := c.triggerAutoMatch(ctx, lastShokoUnknownFileTimestamp); {
		case err == nil:
			lastShokoUnknownFileTimestamp = *nextTimestamp
		case errors.Is(err, shoko.ErrForbidden):
			log.Error().Msg("shoko is unauthorized, routine will stop")
			return
		default:
			log.Error().Msgf("shoko integration failed: %s", err)
		}

		select {
		case <-ticker.C:
		case <-ctx.Done():
			return
		case <-shutdown:
			return
		}
	}
}

func (c *Controller) triggerAutoMatch(
	ctx context.Context,
	lastTimestamp time.Time,
) (*time.Time, error) {
	files, err := c.dep.Shoko.ListUnknownFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing shoko unmatched files: %w", err)
	}

	files = sliceutils.Filter(files, func(f shoko.File) bool {
		return f.CreatedAt.After(lastTimestamp)
	})

	if len(files) == 0 {
		return &lastTimestamp, nil
	}

	list, err := c.dep.AnimeListSource.GetCurrentlyWatching(ctx)
	if err != nil {
		return nil, fmt.Errorf("retrieving watchlist: %w", err)
	}

outer:
	for _, file := range files {
		filename := path.Base(file.RelativePath)
		metadata := metadata.Parse(filename, 1, nil)

		for _, entry := range list {
			for _, title := range entry.Titles {
				if metadata.PrimaryTitle != title {
					continue
				}

				anidbID, err := c.dep.AnidbIDResolver.FromAnilist(ctx, entry.AnilistID)
				if err != nil {
					log.
						Ctx(ctx).
						Error().
						Str("path", file.RelativePath).
						Msg("could not cross-correlate anilist-id to anidb-id")
					continue outer
				}

				// Usually one file per episode, so lastEpisode is nota problem.
				epID, err := c.dep.AnidbEpIDResolver.EpisodeID(ctx, anidbID, int(metadata.Tags.LastEpisode()))
				if err != nil {
					log.
						Ctx(ctx).
						Error().
						Str("path", file.RelativePath).
						Msg("could not cross-correlate anilist episode id to anidb-id")
					continue outer
				}

				matched, err := c.dep.Shoko.MatchFileToCrossReference(ctx, file.ID, anidbID, epID)
				if err != nil {
					return nil, fmt.Errorf("matching unrecognized shoko file: %w", err)
				}

				if matched {
					log.
						Ctx(ctx).
						Info().
						Str("path", file.RelativePath).
						Msg("matched unrecognized shoko file to show")
				} else {
					log.
						Ctx(ctx).
						Info().
						Str("path", file.RelativePath).
						Msg("could not match unrecognized shoko file")
				}

				lastTimestamp = file.CreatedAt
				continue outer
			}
		}

		log.
			Ctx(ctx).
			Info().
			Str("path", file.RelativePath).
			Msg("could not match unrecognized shoko file to any anime in watchlist, skipping")

		lastTimestamp = file.CreatedAt
	}

	return &lastTimestamp, nil
}
