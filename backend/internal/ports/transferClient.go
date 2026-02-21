package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain/transfer"
)

type (
	TransferClientController interface {
		Download(ctx context.Context, releaseCandidate transfer.ReleaseCandidate) (*transfer.ReleaseDownload, error)
	}

	TransferClientControllerFactory interface {
		New(ctx context.Context, client *transfer.Client) (TransferClientController, error)
	}
)
