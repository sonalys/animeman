package nyaa

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type (
	SourceType int
	Query      string
	OrQuery    []string
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
	} else {
		query.Set("q", prevQuery+" "+string(q))
	}
	req.URL.RawQuery = query.Encode()
}

func filterNotEmpty(entries []string) []string {
	notEmpty := make([]string, 0, len(entries))
	for i := range entries {
		if entries[i] != "" {
			notEmpty = append(notEmpty, entries[i])
		}
	}
	return notEmpty
}

func quoteEntriesWithSpace(entries []string) []string {
	out := make([]string, 0, len(entries))
	for i := range entries {
		if strings.Contains(entries[i], " ") {
			out = append(out, fmt.Sprintf("\"%s\"", entries[i]))
			continue
		}
		out = append(out, entries[i])
	}
	return out
}

func (entries OrQuery) Apply(req *http.Request) {
	query := req.URL.Query()
	prevQuery := query.Get("q")

	entries = filterNotEmpty(entries)
	entries = quoteEntriesWithSpace(entries)
	curQuery := fmt.Sprintf("(%s)", strings.Join(entries, "|"))

	if prevQuery == "" {
		query.Set("q", curQuery)
	} else {
		query.Set("q", prevQuery+" "+curQuery)
	}
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
