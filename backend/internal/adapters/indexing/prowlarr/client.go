package prowlarr

import (
	"context"
	"errors"
	"fmt"
	"net/http"
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

		if err, ok := errors.AsType[*starr.ReqError](err); ok {
			if err.Code == http.StatusUnauthorized {
				return "", apperr.New(err, codes.Unauthenticated, "authenticating with prowlarr")
			}
		}

		return "", fmt.Errorf("getting system status: %w", err)
	}

	return status.Version, nil
}
