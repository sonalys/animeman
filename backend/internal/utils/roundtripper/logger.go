package roundtripper

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type loggerTransport struct {
	next http.RoundTripper
}

func (l *loggerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t1 := time.Now()
	ctx := req.Context()

	log.
		Debug().
		Ctx(ctx).
		Stringer("url", req.URL).
		Str("method", req.Method).
		Any("headers", req.Header).
		Msg("Sending request")

	resp, err := l.next.RoundTrip(req)

	if resp != nil {
		log.
			Debug().
			Ctx(ctx).
			Dur("dur", time.Since(t1)).
			Any("headers", resp.Header).
			Int("statusCode", resp.StatusCode).
			Msg("Received response")
	}

	return resp, err
}

func NewLoggerTransport(wrap http.RoundTripper) http.RoundTripper {
	return &loggerTransport{
		next: wrap,
	}
}
