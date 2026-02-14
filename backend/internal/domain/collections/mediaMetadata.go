package collections

import (
	"slices"
	"time"
)

type MediaMetadata struct {
	Genres          []string
	AiringStartedAt time.Time
	AiringEndedAt   time.Time
}

func NewMediaMetadata(
	genres []string,
	airingStartedAt time.Time,
	airingEndedAt time.Time,
) MediaMetadata {
	return MediaMetadata{
		Genres:          slices.Compact(genres),
		AiringStartedAt: airingStartedAt,
		AiringEndedAt:   airingEndedAt,
	}
}
