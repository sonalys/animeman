package collections

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	SeasonID struct{ uuid.UUID }

	Season struct {
		ID      SeasonID
		MediaID MediaID

		Number       int
		AiringStatus AiringStatus

		SeasonMetadata SeasonMetadata

		Episodes []*Episode
	}
)

func (s *Season) NewEpisode(
	t MediaType,
	number string,
	titles []Title,
	airingDate *time.Time,
) *Episode {
	episode := &Episode{
		ID:         shared.NewID[EpisodeID](),
		SeasonID:   s.ID,
		MediaID:    s.MediaID,
		Type:       t,
		Number:     number,
		Titles:     titles,
		AiringDate: airingDate,
	}
	s.Episodes = append(s.Episodes, episode)

	return episode
}
