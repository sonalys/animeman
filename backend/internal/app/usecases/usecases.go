package usecases

import (
	"github.com/sonalys/animeman/internal/ports"
)

type (
	Repositories struct {
		ports.UserRepository
		ports.IndexerClientRepository
		ports.TransferClientRepository
		ports.CollectionRepository
		ports.WatchlistRepository
	}

	usecases struct {
		repositories Repositories
	}
)

func NewUsecases(r Repositories) usecases {
	return usecases{
		repositories: r,
	}
}
