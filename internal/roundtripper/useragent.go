package roundtripper

import (
	"net/http"
)

// throttledTransport Rate Limited HTTP Client
type userAgentTransport struct {
	roundTripperWrap http.RoundTripper
	userAgent        string
}

func (t *userAgentTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	r.Header.Add("User-Agent", t.userAgent)
	return t.roundTripperWrap.RoundTrip(r)
}

// NewRateLimitedTransport wraps transportWrap with a rate limitter
func NewUserAgentTransport(wrap http.RoundTripper, userAgent string) http.RoundTripper {
	return &userAgentTransport{
		roundTripperWrap: wrap,
		userAgent:        userAgent,
	}
}
