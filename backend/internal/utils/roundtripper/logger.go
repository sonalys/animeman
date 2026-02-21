package roundtripper

import (
	"net/http"
	"time"

	"github.com/rs/zerolog/log"
)

type loggerTransport struct {
	wrap http.RoundTripper
}

func (l *loggerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	t1 := time.Now()
	ctx := req.Context()

	log.
		Trace().
		Ctx(ctx).
		Str("url", req.URL.String()).
		Str("method", req.Method).
		Msg("Sending request")

	resp, err := l.wrap.RoundTrip(req)

	if resp != nil {
		log.
			Trace().
			Ctx(ctx).
			Dur("dur", time.Since(t1)).
			Int("statusCode", resp.StatusCode).
			Msg("Received response")
	}

	return resp, err
}

func NewLoggerTransport(wrap http.RoundTripper) http.RoundTripper {
	return &loggerTransport{
		wrap: wrap,
	}
}
