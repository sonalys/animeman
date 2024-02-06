package nyaa

import (
	"net/http"
	"time"

	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://nyaa.si/?page=rss"

type (
	API struct {
		client *http.Client
	}

	Entry struct {
		Title string `xml:"title"`
		Link  string `xml:"link"`
	}

	RSS struct {
		Channel struct {
			Entries []Entry `xml:"item"`
		} `xml:"channel"`
	}

	ListArg interface {
		Apply(*http.Request)
	}
)

func New() *API {
	return &API{
		client: &http.Client{
			Jar: http.DefaultClient.Jar,
			Transport: roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(1*time.Second), 1),
			),
			// Timeout: 10 * time.Second,
		},
	}
}
