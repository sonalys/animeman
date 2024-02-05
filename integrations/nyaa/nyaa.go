package nyaa

import (
	"context"
	"encoding/xml"
	"fmt"
	"net/http"
	"net/url"
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

	SourceType int
	Query      string
	Category   string
	User       string
)

const (
	SourceTypeAll SourceType = iota
	SourceTypeNoRemake
	SourceTypeTrusted
)

const (
	CategoryAnime                  Category = "1_0"
	CategoryAnimeEnglishTranslated Category = "1_2"
)

func New() *API {
	return &API{
		client: &http.Client{
			Transport: roundtripper.NewRateLimitedTransport(
				http.DefaultTransport, rate.NewLimiter(rate.Every(60*time.Second), 60),
			),
			Timeout: 3 * time.Second,
		},
	}
}

func (s SourceType) Apply(req *http.Request) {
	q := req.URL.Query()
	q.Add("f", fmt.Sprint(s))
	req.URL.RawQuery = q.Encode()
}

func (q Query) Apply(req *http.Request) {
	query := req.URL.Query()
	prevQuery := query.Get("q")
	if prevQuery == "" {
		query.Set("q", string(q))
	}
	query.Set("q", prevQuery+" "+string(q))
	req.URL.RawQuery = query.Encode()
}

func (c Category) Apply(req *http.Request) {
	query := req.URL.Query()
	query.Add("c", string(c))
	req.URL.RawQuery = query.Encode()
}

func (u User) Apply(req *http.Request) {
	query := req.URL.Query()
	query.Add("u", url.QueryEscape(string(u)))
	req.URL.RawQuery = query.Encode()
}

func (api *API) List(ctx context.Context, args ...ListArg) ([]Entry, error) {
	var path = API_URL
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, path, nil)
	if err != nil {
		return nil, fmt.Errorf("creating request: %w", err)
	}
	for _, f := range args {
		f.Apply(req)
	}
	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()
	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}
	return feed.Channel.Entries, nil
}
