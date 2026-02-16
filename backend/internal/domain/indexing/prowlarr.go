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

var indexerTypeStringer = map[IndexerType]string{
	IndexerTypeProwlarr: "prowlarr",
}

func NewClient(
	userID shared.UserID,
	t IndexerType,
	address url.URL,
	auth authentication.Authentication,
) *IndexerClient {
	return &IndexerClient{
		ID:             shared.NewID[IndexerID](),
		OwnerID:        userID,
		Type:           t,
		Address:        address,
		Authentication: auth,
	}
}

func (s IndexerType) String() string {
	if value, ok := indexerTypeStringer[s]; ok {
		return value
	}
	return ""
}

func (s IndexerType) IsValid() bool {
	return s > IndexerTypeUnknown && s < indexerTypeSentinel
}
