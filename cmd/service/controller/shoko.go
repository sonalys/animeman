package controller

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/ports/shoko"
)

func (c *Controller) startShokoRoutine(ctx context.Context, shutdown <-chan struct{}) {
	ticker := time.NewTicker(time.Minute)
	defer ticker.Stop()

	var lastShokoUnknownFileTimestamp time.Time

	for {
		switch nextTimestamp, err := c.TriggerAutoMatch(ctx, lastShokoUnknownFileTimestamp); {
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

func (c *Controller) TriggerAutoMatch(
	ctx context.Context,
	lastTimestamp time.Time,
) (*time.Time, error) {
	files, err := c.dep.Shoko.ListUnknownFiles(ctx)
	if err != nil {
		return nil, fmt.Errorf("listing shoko unmatched files: %w", err)
	}

	for _, file := range files {
		if file.CreatedAt.Before(lastTimestamp) {
			continue
		}

		matched, err := c.dep.Shoko.AutoMatchAndSaveFile(ctx, file.ID)
		if err != nil {
			return nil, fmt.Errorf("auto-matching file: %w", err)
		}

		if matched {
			log.
				Ctx(ctx).
				Info().
				Str("path", file.RelativePath).
				Msg("matched shoko file")
		} else {
			log.
				Ctx(ctx).
				Info().
				Str("path", file.RelativePath).
				Msg("could not match shoko file")
		}

		lastTimestamp = file.CreatedAt
	}

	return &lastTimestamp, nil
}
