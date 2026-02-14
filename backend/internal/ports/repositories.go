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
	UpdateHandler[T any] = func(*T) error

	ListOptions struct {
		PageSize int32
		Cursor   shared.ID
	}

	UserRepository interface {
		Create(ctx context.Context, user *users.User) error
		Get(ctx context.Context, id shared.UserID) (*users.User, error)
		Update(ctx context.Context, id shared.UserID, updateHandler UpdateHandler[users.User]) error
		Delete(ctx context.Context, id shared.UserID) error
	}

	IndexerClientRepository interface {
		Create(ctx context.Context, client *indexing.IndexerClient) error
		ListByOwner(ctx context.Context, id shared.UserID) ([]indexing.IndexerClient, error)
		Update(ctx context.Context, id indexing.IndexerID, updateHandler UpdateHandler[indexing.IndexerClient]) error
		Delete(ctx context.Context, id indexing.IndexerID) error
	}

	TransferClientRepository interface {
		Create(ctx context.Context, client *transfer.Client) error
		ListByOwner(ctx context.Context, id shared.UserID) ([]transfer.Client, error)
		Update(ctx context.Context, id transfer.ClientID, updateHandler UpdateHandler[transfer.Client]) error
		Delete(ctx context.Context, id transfer.ClientID) error
	}

	CollectionRepository interface {
		Create(ctx context.Context, collection *collections.Collection) error
		ListByOwner(ctx context.Context, id shared.UserID) ([]collections.Collection, error)
		Update(ctx context.Context, id collections.CollectionID, updateHandler UpdateHandler[collections.Collection]) error
		Delete(ctx context.Context, id collections.CollectionID) error
	}

	QualityProfileRepository interface {
		Create(ctx context.Context, qualityProfile *collections.QualityProfile) error
		List(ctx context.Context) ([]collections.QualityProfile, error)
		Update(ctx context.Context, id collections.QualityProfileID, updateHandler UpdateHandler[collections.QualityProfile]) error
		Delete(ctx context.Context, id collections.QualityProfileID) error
	}

	MediaRepository interface {
		Create(ctx context.Context, media *collections.Media) error
		ListByCollection(ctx context.Context, id collections.CollectionID, opts ListOptions) ([]collections.Media, error)
		Update(ctx context.Context, id collections.MediaID, updateHandler UpdateHandler[collections.Media]) error
		Delete(ctx context.Context, id collections.MediaID) error
	}
)
