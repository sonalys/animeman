package anilist

import (
	"net/http"
	"time"

	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://graphql.anilist.co"
const userAgent = "github.com/sonalys/animeman"

type (
	API struct {
		Username string
		client   *http.Client
	}
)

func New(username string) *API {
	client := &http.Client{
		Transport: roundtripper.NewUserAgentTransport(
			roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(time.Second), 1),
			), userAgent),
		Timeout: 10 * time.Second,
	}
	return &API{
		client:   client,
		Username: username,
	}
}
