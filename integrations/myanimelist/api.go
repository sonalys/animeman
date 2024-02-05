package myanimelist

import (
	"net/http"
	"time"

	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://myanimelist.net"

type (
	API struct {
		Username string
		client   *http.Client
	}
)

func New(username string) *API {
	return &API{
		Username: username,
		client: &http.Client{
			Transport: roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(time.Second), 1),
			),
			Timeout: 10 * time.Second,
		},
	}
}
