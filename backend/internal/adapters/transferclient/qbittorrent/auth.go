package qbittorrent

import (
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	"github.com/sonalys/animeman/internal/app/apperr"
	"google.golang.org/grpc/codes"
)

type authTransport struct {
	next     http.RoundTripper
	username string
	password string
}

func newAuthTransport(next http.RoundTripper, username, password string) http.RoundTripper {
	return &authTransport{
		next:     next,
		username: username,
		password: password,
	}
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	ctx := req.Context()

	for {
		resp, err := t.next.RoundTrip(req.Clone(ctx))
		if err != nil || resp.StatusCode != http.StatusForbidden {
			return resp, err
		}

		// Drain the body.
		_, _ = io.Copy(io.Discard, resp.Body)

		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("closing response body: %w", err)
		}

		var path = "/auth/login"

		body := url.Values{
			"username": {t.username},
			"password": {t.password},
		}

		authReq, err := http.NewRequestWithContext(ctx, http.MethodPost, path, strings.NewReader(body.Encode()))
		if err != nil {
			return nil, fmt.Errorf("creating login request: %w", err)
		}

		authReq.Header.Add("Content-Type", "application/x-www-form-urlencoded")

		authResp, err := t.next.RoundTrip(authReq)
		if err != nil {
			return nil, apperr.New(err, codes.Unauthenticated, "authenticating with qbittorrent")
		}

		authRespBody, err := io.ReadAll(authResp.Body)
		if err != nil {
			_ = resp.Body.Close()
			return nil, fmt.Errorf("reading authentication response body")
		}

		if err := resp.Body.Close(); err != nil {
			return nil, fmt.Errorf("closing response body: %w", err)
		}

		if resp.StatusCode >= 400 {
			return nil, fmt.Errorf("authenticating: %s", string(authRespBody))
		}
	}
}
