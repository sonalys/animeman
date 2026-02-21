package transfer

import (
	"net/url"

	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type ClientBuilder struct {
	Owner      shared.UserID
	ClientType ClientType
	Address    url.URL
	Auth       authentication.Authentication
}

func NewClientBuilder() *ClientBuilder {
	return &ClientBuilder{
		ClientType: ClientTypeUnknown,
	}
}

func (b *ClientBuilder) WithOwner(id shared.UserID) *ClientBuilder {
	b.Owner = id
	return b
}

func (b *ClientBuilder) WithType(t ClientType) *ClientBuilder {
	b.ClientType = t
	return b
}

func (b *ClientBuilder) WithAddress(addr url.URL) *ClientBuilder {
	b.Address = addr
	return b
}

func (b *ClientBuilder) WithAuth(auth authentication.Authentication) *ClientBuilder {
	b.Auth = auth
	return b
}

func (b *ClientBuilder) Build() (*Client, error) {
	return NewClient(b.Owner, b.ClientType, b.Address, b.Auth)
}
