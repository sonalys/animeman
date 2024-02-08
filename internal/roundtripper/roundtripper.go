package roundtripper

import (
	"net/http"

	"golang.org/x/time/rate"
)

// throttledTransport Rate Limited HTTP Client
type throttledTransport struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      *rate.Limiter
}

func (t *throttledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	err := t.ratelimiter.Wait(r.Context()) // This is a blocking call. Honors the rate limit
	if err != nil {
		return nil, err
	}
	return t.roundTripperWrap.RoundTrip(r)
}

// NewRateLimitedTransport wraps transportWrap with a rate limitter
func NewRateLimitedTransport(wrap http.RoundTripper, limiter *rate.Limiter) http.RoundTripper {
	return &throttledTransport{
		roundTripperWrap: wrap,
		ratelimiter:      limiter,
	}
}
