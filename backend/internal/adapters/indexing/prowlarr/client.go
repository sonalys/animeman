package prowlarr

import (
	"context"
	"errors"
	"fmt"
	"net/url"
	"syscall"

	"github.com/sonalys/animeman/internal/app/apperr"
	"golift.io/starr"
	"golift.io/starr/prowlarr"
	"google.golang.org/grpc/codes"
)

type (
	Client struct {
		wrapper *prowlarr.Prowlarr
	}
)

func New(host url.URL, apiKey string) *Client {
	wrapper := prowlarr.New(&starr.Config{
		URL:    host.String(),
		APIKey: apiKey,
	})

	return &Client{
		wrapper: wrapper,
	}
}

func (c *Client) Version(ctx context.Context) (string, error) {
	status, err := c.wrapper.GetSystemStatusContext(ctx)
	if err != nil {
		if errors.Is(err, syscall.ECONNREFUSED) {
			return "", apperr.New(err, codes.InvalidArgument, "hostname refused the connection")
		}

		return "", fmt.Errorf("getting system status: %w", err)
	}

	return status.Version, nil
}
