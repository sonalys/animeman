package nyaa

import (
	"fmt"
	"net/http"
	"net/url"
)

type (
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
