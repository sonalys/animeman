package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain/indexing"
)

type (
	IndexingClientController interface {
	}

	IndexingClientControllerFactory interface {
		New(ctx context.Context, client *indexing.Client) (IndexingClientController, error)
	}
)
