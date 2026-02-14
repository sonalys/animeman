package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type CollectionID struct{ uuid.UUID }

type Collection struct {
	ID       CollectionID
	Name     string
	BasePath string
	Tags     []string

	Monitored bool

	CreatedAt time.Time
}
