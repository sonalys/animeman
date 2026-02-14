package indexing

import (
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	IndexerID   struct{ uuid.UUID }
	IndexerType uint

	IndexerClient struct {
		ID      IndexerID
		OwnerID shared.UserID

		Type           IndexerType
		Address        url.URL
		Authentication authentication.Authentication
	}
)

const (
	IndexerTypeUnknown IndexerType = iota
	IndexerTypeProwlarr
	indexerTypeSentinel
)

var IndexerTypeStringer = map[IndexerType]string{
	IndexerTypeProwlarr: "prowlarr",
}

func (s IndexerType) String() string {
	if value, ok := IndexerTypeStringer[s]; ok {
		return value
	}
	return ""
}

func (s IndexerType) IsValid() bool {
	return s > IndexerTypeUnknown && s < indexerTypeSentinel
}
