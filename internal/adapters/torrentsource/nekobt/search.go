package nekobt

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
)

type (
	rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel  `xml:"channel"`
	}

	channel struct {
		Title string `xml:"title"`
		Items []item `xml:"item"`
	}

	item struct {
		Title   string `xml:"title"`
		Link    string `xml:"link"`
		GUID    string `xml:"guid"`
		PubDate string `xml:"pubDate"`
		Size    int64  `xml:"size"`
		Attrs   []attr `xml:"attr"`
	}

	attr struct {
		Name  string `xml:"name,attr"`
		Value string `xml:"value,attr"`
	}
)

// Search implements torrentsource.Source.
// nekoBT matches the entry by AniList/MAL id directly, so no title filtering is needed.
func (api *API) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, TORZNAB_URL, nil))

	q, err := api.buildQuery(entry, opts)
	if err != nil {
		return nil, fmt.Errorf("building query: %w", err)
	}

	req.URL.RawQuery = q

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var feed rss
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	items := utils.Filter(feed.Channel.Items,
		filterSeeders(1),
		filterMetadata(entry),
	)

	torrents := utils.Map(items, func(item item) torrentsource.Torrent {
		return torrentsource.Torrent{
			Title:   item.Title,
			Link:    item.Link,
			Seeders: item.seeders(),
			Hash:    item.attr("infohash"),
		}
	})

	torrents = parser.Prioritize(entry, torrents, opts)

	log.
		Ctx(ctx).
		Debug().
		Int("results", len(torrents)).
		Msg("search results")

	return torrents, nil
}

func (api *API) buildQuery(entry animelist.Entry, opt torrentsource.SearchOptions) (string, error) {
	q := url.Values{}

	q.Set("t", "search")

	mediaID, err := resolveMediaID(entry)
	if err != nil {
		return "", err
	}

	q.Set("media_id", mediaID)

	// media_id already narrows to the entry, so `q` only carries the
	// quality/source filters and the user suffix, same format as nyaa.
	if query := buildQuery(opt); query != "" {
		q.Set("q", query)
	}

	if api.config.APIKey != "" {
		q.Set("apikey", api.config.APIKey)
	}

	// Apply user-configured torznab parameters last, so they can
	// override anything built above (sort, sub_lang, mtl, hardsub, ...).
	for name, value := range api.config.CustomParameters {
		q.Set(name, value)
	}

	return q.Encode(), nil
}

var querySanitization = strings.NewReplacer(
	"\"", " ",
	"-", " ",
	"\"", " ",
	"'", " ",
	"(", " ",
	")", " ",
)

// buildQuery builds the `q` search query from the search options,
// mirroring the nyaa adapter: qualities, sources and the user suffix.
func buildQuery(opt torrentsource.SearchOptions) string {
	var parts []string

	if len(opt.Qualities) > 0 {
		var b strings.Builder
		b.WriteString("(")
		for i, quality := range opt.Qualities {
			if i > 0 {
				b.WriteString("|")
			}
			b.WriteString(strconv.Quote(querySanitization.Replace(quality)))
		}
		b.WriteString(")")
		parts = append(parts, b.String())
	}

	if len(opt.Sources) > 0 {
		var b strings.Builder
		b.WriteString("(")
		for i, source := range opt.Sources {
			if i > 0 {
				b.WriteString("|")
			}
			b.WriteString(strconv.Quote(querySanitization.Replace(source)))
		}
		b.WriteString(")")
		parts = append(parts, b.String())
	}

	if opt.SearchSuffix != "" {
		parts = append(parts, querySanitization.Replace(opt.SearchSuffix))
	}

	return strings.Join(parts, " ")
}

// resolveMediaID returns the nekoBT external id for the entry, e.g. `anilist-20594`.
func resolveMediaID(entry animelist.Entry) (string, error) {
	switch {
	case entry.AnilistID != 0:
		return fmt.Sprintf("anilist-%d", entry.AnilistID), nil
	case entry.MALID != 0:
		return fmt.Sprintf("mal-%d", entry.MALID), nil
	default:
		return "", fmt.Errorf("entry %q has no anilist or mal id", strings.Join(entry.Titles, ", "))
	}
}

func filterSeeders(minSeeders int) func(item) bool {
	return func(item item) bool {
		return item.seeders() >= minSeeders
	}
}

func filterMetadata(entry animelist.Entry) func(item) bool {
	return func(item item) bool {
		// Compare the published date of the torrent with the entry's start and end dates.
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			return false
		}

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if !entry.StartDate.IsZero() && pubDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		return true
	}
}

func (i item) attr(name string) string {
	for _, a := range i.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func (i item) seeders() int {
	v, err := strconv.Atoi(i.attr("seeders"))
	if err != nil {
		return 0
	}
	return v
}
