package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain/transfer"
)

type (
	TransferClientController interface {
		Wait(ctx context.Context)
		Download(ctx context.Context, releaseCandidate transfer.ReleaseCandidate) (*transfer.ReleaseDownload, error)
	}
)
