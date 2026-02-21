package indexing

import (
	"context"
	"fmt"

	"github.com/sonalys/animeman/internal/adapters/indexing/prowlarr"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/ports"
)

type (
	factory struct{}
)

func NewFactory() factory {
	return factory{}
}

func (f factory) New(ctx context.Context, client *indexing.Client) (ports.IndexingClientController, error) {
	switch client.Type {
	case indexing.IndexerTypeProwlarr:
		auth, ok := client.Authentication.AsAPIKey()
		if !ok {
			return nil, fmt.Errorf("unsupported authentication type: %s", client.Authentication.Type)
		}

		client := prowlarr.New(client.Address, auth.Key)

		if _, err := client.Version(ctx); err != nil {
			return nil, fmt.Errorf("probing prowlarr readiness: %w", err)
		}

		return client, nil
	default:
		return nil, fmt.Errorf("unexpected client type: %s", client.Type)
	}
}
