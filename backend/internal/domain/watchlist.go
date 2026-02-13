package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	WatchlistID      struct{ uuid.UUID }
	WatchlistEntryID struct{ uuid.UUID }

	Watchlist struct {
		ID     WatchlistID
		UserID UserID

		Source     WatchlistSource
		ExternalID string

		SyncFrequency time.Duration
		LastSyncedAt  time.Time

		CreatedAt time.Time
	}

	WatchlistEntry struct {
		ID          WatchlistEntryID
		WatchlistID WatchlistID
		UserID      UserID

		MediaID     MediaID
		SeasonID    SeasonID
		LastWatched EpisodeID

		Status WatchlistStatus

		CreatedAt time.Time
		UpdatedAt time.Time
	}
)
