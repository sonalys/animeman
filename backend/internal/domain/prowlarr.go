package domain

import "github.com/gofrs/uuid/v5"

type (
	ProwlarrConfiguration struct {
		ID      uuid.UUID
		OwnerID uuid.UUID
		Host    string
		APIKey  string
	}
)
