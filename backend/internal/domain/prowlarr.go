package domain

import "github.com/gofrs/uuid/v5"

type (
	ProwlarrConfigID struct{ uuid.UUID }

	ProwlarrConfiguration struct {
		ID      ProwlarrConfigID
		OwnerID UserID
		Host    string
		APIKey  string
	}
)
