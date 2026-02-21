package qbittorrent

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type authTransport struct {
	transport http.RoundTripper
	username  string
	password  string
}

func (t *authTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqClone := req.Clone(req.Context())

	resp, err := t.transport.RoundTrip(reqClone)
	if err != nil || resp.StatusCode != http.StatusForbidden {
		return resp, err
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	var path = "/auth/login"

	data := url.Values{
		"username": {t.username},
		"password": {t.password},
	}.Encode()

	req, err = http.NewRequest(http.MethodPost, path, strings.NewReader(data))
	if err != nil {
		return nil, fmt.Errorf("creating login request: %w", err)
	}

	req.Header.Add("Content-Type", "application/x-www-form-urlencoded")

	resp, err = t.transport.RoundTrip(req)
	if err != nil {
		return nil, fmt.Errorf("logging in: %w", err)
	}

	if err := resp.Body.Close(); err != nil {
		return nil, err
	}

	reqClone = req.Clone(req.Context())

	return t.transport.RoundTrip(reqClone)
}
