package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	EpisodeID struct{ uuid.UUID }

	Episode struct {
		ID       EpisodeID
		SeasonID SeasonID
		MediaID  MediaID
		// Number is a string due to complex episode number variations, like episode 6.5 or combined episodes.
		Number     string
		Title      string
		AiringDate *time.Time
		Files      []LocalFile
	}
)
