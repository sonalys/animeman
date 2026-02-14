package collections

import (
	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	ProwlarrConfigID struct{ uuid.UUID }

	ProwlarrConfiguration struct {
		ID      ProwlarrConfigID
		OwnerID shared.UserID
		Host    string
		APIKey  string
	}
)
