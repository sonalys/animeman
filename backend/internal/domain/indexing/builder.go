package indexing

import (
	"net/url"

	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type ClientBuilder struct {
	id         IndexerID
	ownerID    shared.UserID
	clientType ClientType
	address    url.URL
	auth       authentication.Authentication
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		clientType: IndexerTypeUnknown,
	}
}

func (b *ClientBuilder) WithID(id IndexerID) *ClientBuilder {
	b.id = id
	return b
}

func (b *ClientBuilder) WithOwner(id shared.UserID) *ClientBuilder {
	b.ownerID = id
	return b
}

func (b *ClientBuilder) WithType(t ClientType) *ClientBuilder {
	b.clientType = t
	return b
}

func (b *ClientBuilder) WithAddress(addr url.URL) *ClientBuilder {
	b.address = addr
	return b
}

func (b *ClientBuilder) WithAuth(auth authentication.Authentication) *ClientBuilder {
	b.auth = auth
	return b
}

func (b *ClientBuilder) Build() (*Client, error) {
	return NewClient(b.ownerID, b.clientType, b.address, b.auth), nil
}
