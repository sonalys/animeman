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
	// 1. Clone the request to avoid mutating the original (thread safety)
	newReq := req.Clone(req.Context())

	// 2. Override the Host and Scheme
	newReq.URL.Host = t.baseURL.Host
	newReq.URL.Scheme = t.baseURL.Scheme

	// 3. Prepend the Base Path to the existing request path
	// Example: Base "/api/v1" + Req "/users" => "/api/v1/users"
	newReq.URL.Path = path.Join(t.baseURL.Path, req.URL.Path)

	// 4. Pass the modified request to the underlying transport
	return t.base.RoundTrip(newReq)
}
