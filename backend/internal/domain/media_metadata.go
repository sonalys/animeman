package domain

import "time"

type MediaMetadata struct {
	Genres          []string
	AiringStartedAt time.Time
	AiringEndedAt   time.Time
}
