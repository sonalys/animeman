package nyaa

import (
	"net/http"
	"time"

	"github.com/sonalys/animeman/internal/roundtripper"
	"golang.org/x/time/rate"
)

const API_URL = "https://nyaa.si/?page=rss"

type (
	Config struct {
		ListParameters map[string]string
	}

	API struct {
		config Config
		client *http.Client
	}

	Entry struct {
		Title   string `xml:"title"`
		Link    string `xml:"link"`
		PubDate string `xml:"pubDate"`
		Seeders int    `xml:"nyaa:seeders"`
	}

	RSS struct {
		Channel struct {
			Entries []Entry `xml:"item"`
		} `xml:"channel"`
	}
)

func New(c Config) *API {
	return &API{
		config: c,
		client: &http.Client{
			Jar: http.DefaultClient.Jar,
			Transport: roundtripper.NewUserAgentTransport(roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(1*time.Second), 1),
			), "github.com/sonalys/animeman"),
			// Timeout: 10 * time.Second,
		},
	}
}
