package mappers

import (
	"net/url"

	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/utils/errutils"
)

func NewIndexerClient(from *sqlcgen.ProwlarrConfiguration) *indexing.IndexerClient {
	return &indexing.IndexerClient{
		ID:             shared.ParseID[indexing.IndexerID](from.ID),
		OwnerID:        shared.ParseID[shared.UserID](from.OwnerID),
		Address:        *errutils.Must(url.Parse(from.Host)),
		Authentication: authentication.NewAPIKeyAuthentication(from.ApiKey),
	}
}
