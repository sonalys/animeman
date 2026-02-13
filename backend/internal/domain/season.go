package domain

import "github.com/gofrs/uuid/v5"

type (
	SeasonID struct{ uuid.UUID }

	Season struct {
		ID      SeasonID
		MediaID MediaID

		Number       int
		AiringStatus AiringStatus

		SeasonMetadata SeasonMetadata

		Episodes []Episode
	}
)
