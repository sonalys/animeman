package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain"
)

func NewProwlarrConfiguration(from *sqlcgen.ProwlarrConfiguration) *domain.ProwlarrConfiguration {
	return &domain.ProwlarrConfiguration{
		ID:      domain.ParseID[domain.ProwlarrConfigID](from.ID),
		OwnerID: domain.ParseID[domain.UserID](from.OwnerID),
		Host:    from.Host,
		APIKey:  from.ApiKey,
	}
}
