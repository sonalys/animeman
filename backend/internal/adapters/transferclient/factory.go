package transferclient

import (
	"context"
	"fmt"

	"github.com/sonalys/animeman/internal/adapters/transferclient/qbittorrent"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/ports"
)

type (
	factory struct{}
)

func NewFactory() factory {
	return factory{}
}

func (f factory) New(ctx context.Context, client *transfer.Client) (ports.TransferClientController, error) {
	switch client.Type {
	case transfer.ClientTypeQBittorrent:
		auth, ok := client.Authentication.AsUserPassword()
		if !ok {
			return nil, fmt.Errorf("unsupported authentication type: %s", client.Authentication.Type)
		}

		controller, err := qbittorrent.New(ctx, client.Address.String(), auth.Username, string(auth.Password))
		if err != nil {
			return nil, fmt.Errorf("initializing qbittorrent controller: %w", err)
		}

		if _, err := controller.Version(ctx); err != nil {
			return nil, fmt.Errorf("probing qbittorrent readiness: %w", err)
		}

		return controller, nil
	default:
		return nil, fmt.Errorf("unexpected client type: %s", client.Type)
	}
}
