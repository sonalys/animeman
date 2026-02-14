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
		Titles     []AlternativeTitle
		AiringDate *time.Time
		Files      []CollectionFile
	}
)

func (e *Episode) BestFile() *CollectionFile {
	if len(e.Files) == 0 {
		return nil
	}
	best := &e.Files[0]
	for i := 1; i < len(e.Files); i++ {
		if e.Files[i].Quality.Resolution > best.Quality.Resolution {
			best = &e.Files[i]
			continue
		}

		if e.Files[i].Quality.Codec > best.Quality.Codec {
			best = &e.Files[i]
			continue
		}
	}
	return best
}
