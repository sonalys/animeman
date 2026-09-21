package roundtripper

import (
	"context"
	"net/http"
	"strconv"
	"sync"
	"time"

	"golang.org/x/time/rate"
)

// throttledTransport Rate Limited HTTP Client.
type throttledTransport struct {
	roundTripperWrap http.RoundTripper
	ratelimiter      *rate.Limiter

	mu         sync.Mutex
	blockUntil time.Time // server-imposed cooldown from rate limit headers
}

func (t *throttledTransport) RoundTrip(r *http.Request) (*http.Response, error) {
	err := t.ratelimiter.Wait(r.Context()) // This is a blocking call. Honors the rate limit.
	if err != nil {
		return nil, err
	}

	t.mu.Lock()
	delay := time.Until(t.blockUntil)
	t.mu.Unlock()

	if delay > 0 {
		if err := wait(r.Context(), delay); err != nil {
			return nil, err
		}
	}

	resp, err := t.roundTripperWrap.RoundTrip(r)
	if err != nil {
		return nil, err
	}
	t.updateDelay(resp)
	return resp, nil
}

// updateDelay records the cooldown announced by the server's rate limit headers.
func (t *throttledTransport) updateDelay(resp *http.Response) {
	delay, ok := GetRetryDelay(resp, time.Now())
	if !ok || delay <= 0 {
		return
	}
	t.mu.Lock()
	if until := time.Now().Add(delay); until.After(t.blockUntil) {
		t.blockUntil = until
	}
	t.mu.Unlock()
}

// GetRetryDelay extracts a server-imposed cooldown from common rate limit headers.
func GetRetryDelay(resp *http.Response, now time.Time) (time.Duration, bool) {
	if v := resp.Header.Get("Retry-After"); v != "" {
		if secs, err := strconv.Atoi(v); err == nil {
			return time.Duration(secs) * time.Second, true
		}
		if at, err := http.ParseTime(v); err == nil {
			return at.Sub(now), true
		}
	}
	if remaining, ok := firstHeader(
		resp,
		"RateLimit-Remaining",
		"X-RateLimit-Remaining",
	); !ok ||
		remaining != "0" {
		return 0, false
	}
	reset, ok := firstHeader(resp, "RateLimit-Reset", "X-RateLimit-Reset")
	if !ok {
		return 0, false
	}
	secs, err := strconv.ParseInt(reset, 10, 64)
	if err != nil {
		return 0, false
	}
	at := time.Unix(secs, 0)
	if secs < 1e9 { // delta-seconds instead of epoch
		at = now.Add(time.Duration(secs) * time.Second)
	}
	return at.Sub(now), true
}

func firstHeader(resp *http.Response, keys ...string) (string, bool) {
	for _, k := range keys {
		if v := resp.Header.Get(k); v != "" {
			return v, true
		}
	}
	return "", false
}

func wait(ctx context.Context, d time.Duration) error {
	timer := time.NewTimer(d)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

// NewRateLimitedTransport wraps transportWrap with a rate limitter.
func NewRateLimitedTransport(wrap http.RoundTripper, limiter *rate.Limiter) http.RoundTripper {
	return &throttledTransport{
		roundTripperWrap: wrap,
		ratelimiter:      limiter,
	}
}
