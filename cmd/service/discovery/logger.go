package discovery

import (
	"context"

	"github.com/rs/zerolog"
)

func getLogger(ctx context.Context) zerolog.Logger {
	logger := zerolog.Ctx(ctx)

	return *logger
}
