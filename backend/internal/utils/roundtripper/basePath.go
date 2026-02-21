package roundtripper

import (
	"net/http"
	"net/url"
)

type prefixRoundTripper struct {
	prefix *url.URL
	base   http.RoundTripper
}

func NewPrefix(prefix *url.URL, next http.RoundTripper) http.RoundTripper {
	return &prefixRoundTripper{
		prefix: prefix,
		base:   next,
	}
}

func (t *prefixRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	req = req.Clone(req.Context())

	req.URL = t.prefix.ResolveReference(t.prefix.JoinPath(req.URL.Path))
	req.Host = req.URL.Host

	base := t.base
	if base == nil {
		base = http.DefaultTransport
	}

	return base.RoundTrip(req)
}
