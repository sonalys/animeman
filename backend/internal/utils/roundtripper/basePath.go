package roundtripper

import (
	"net/http"
	"net/url"
	"path"
)

type basePathTransport struct {
	base    http.RoundTripper
	baseURL *url.URL
}

func NewBasePathTransport(wrap http.RoundTripper, basePath *url.URL) http.RoundTripper {
	return &basePathTransport{
		base:    wrap,
		baseURL: basePath,
	}
}

func (t *basePathTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	newReq := req.Clone(req.Context())
	newReq.URL.Host = t.baseURL.Host
	newReq.URL.Scheme = t.baseURL.Scheme
	newReq.URL.Path = path.Join(t.baseURL.Path, req.URL.Path)

	return t.base.RoundTrip(newReq)
}
