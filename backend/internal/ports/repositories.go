package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/users"
)

type (
	UserRepository interface {
		Create(ctx context.Context, user *users.User) error
		Get(ctx context.Context, id shared.UserID) (*users.User, error)
		Update(ctx context.Context, id shared.UserID, update func(user *users.User) error) error
		Delete(ctx context.Context, id shared.UserID) error
	}

	IndexerClientRepository interface {
		Create(ctx context.Context, config *indexing.IndexerClient) error
		GetByOwner(ctx context.Context, owner shared.UserID) (*indexing.IndexerClient, error)
		Update(ctx context.Context, id indexing.IndexerID, update func(config *indexing.IndexerClient) error) error
		Delete(ctx context.Context, id indexing.IndexerID) error
	}
)
