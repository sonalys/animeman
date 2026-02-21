package indexing

import (
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	IndexerID  struct{ uuid.UUID }
	ClientType uint

	Client struct {
		ID      IndexerID
		OwnerID shared.UserID

		Type           ClientType
		Address        url.URL
		Authentication authentication.Authentication
	}
)

const (
	IndexerTypeUnknown ClientType = iota
	IndexerTypeProwlarr
	indexerTypeSentinel
)

var indexerTypeStringer = map[ClientType]string{
	IndexerTypeProwlarr: "prowlarr",
}

func NewClient(
	userID shared.UserID,
	t ClientType,
	address url.URL,
	auth authentication.Authentication,
) *Client {
	return &Client{
		ID:             shared.NewID[IndexerID](),
		OwnerID:        userID,
		Type:           t,
		Address:        address,
		Authentication: auth,
	}
}

func (s ClientType) String() string {
	if value, ok := indexerTypeStringer[s]; ok {
		return value
	}
	return ""
}

func (s ClientType) IsValid() bool {
	return s > IndexerTypeUnknown && s < indexerTypeSentinel
}
