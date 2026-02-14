package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/transfer"
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
		Create(ctx context.Context, client *indexing.IndexerClient) error
		ListByOwner(ctx context.Context, owner shared.UserID) ([]indexing.IndexerClient, error)
		Update(ctx context.Context, id indexing.IndexerID, update func(client *indexing.IndexerClient) error) error
		Delete(ctx context.Context, id indexing.IndexerID) error
	}

	TransferClientRepository interface {
		Create(ctx context.Context, client *transfer.Client) error
		ListByOwner(ctx context.Context, owner shared.UserID) ([]transfer.Client, error)
		Update(ctx context.Context, id transfer.ClientID, update func(client *transfer.Client) error) error
		Delete(ctx context.Context, id transfer.ClientID) error
	}

	CollectionRepository interface {
		Create(ctx context.Context, collection *collections.Collection) error
		ListByOwner(ctx context.Context, owner shared.UserID) ([]collections.Collection, error)
		Update(ctx context.Context, id collections.CollectionID, update func(collection *collections.Collection) error) error
		Delete(ctx context.Context, id collections.CollectionID) error
	}

	QualityProfileRepository interface {
		Create(ctx context.Context, profile *collections.QualityProfile) error
		List(ctx context.Context) ([]collections.QualityProfile, error)
		Update(ctx context.Context, id collections.QualityProfileID, update func(profile *collections.QualityProfile) error) error
		Delete(ctx context.Context, id collections.QualityProfileID) error
	}
)
