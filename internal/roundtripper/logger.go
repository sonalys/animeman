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

	resp, err := l.wrap.RoundTrip(req)

	if resp != nil {
		log.
			Trace().
			Str("url", req.URL.String()).
			Int("status_code", resp.StatusCode).
			Dur("duration", time.Since(t1)).
			Msg("outgoing request")
	}

	return resp, err
}

func NewLoggerTransport(wrap http.RoundTripper) http.RoundTripper {
	return &loggerTransport{}
}
