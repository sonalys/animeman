package roundtripper

import (
	"net/http"
)

type userAgentTransport struct {
	roundTripperWrap http.RoundTripper
	userAgent        string
}

func (t *userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("User-Agent", t.userAgent)
	return t.roundTripperWrap.RoundTrip(r)
}

// NewUserAgentTransport adds an user-agent to all the requests from an http.client.
func NewUserAgentTransport(wrap http.RoundTripper, userAgent string) http.RoundTripper {
	return &userAgentTransport{
		roundTripperWrap: wrap,
		userAgent:        userAgent,
	}
}
