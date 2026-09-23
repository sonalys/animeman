package nyaa

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"slices"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/pkg/must"
	"github.com/sonalys/animeman/internal/pkg/nyaaquerier"
	"github.com/sonalys/animeman/internal/pkg/parser"
	"github.com/sonalys/animeman/internal/pkg/searcher"
	"github.com/sonalys/animeman/internal/pkg/sliceutils"
	"github.com/sonalys/animeman/internal/pkg/stringutils"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
)

const (
	pageSize = 100
)

type (
	// item represents a single torrent entry in the RSS feed
	item struct {
		Title       string `xml:"title"`
		Link        string `xml:"link"`
		GUID        string `xml:"guid"`
		PubDate     string `xml:"pubDate"`
		Description string `xml:"description"`

		// Nyaa specific fields
		// Note: We use the local name (e.g., "seeders") in the tag.
		// The xml parser handles the "nyaa:" prefix automatically by matching the local name.
		Seeders    int    `xml:"seeders"`
		Leechers   int    `xml:"leechers"`
		Downloads  int    `xml:"downloads"`
		InfoHash   string `xml:"infoHash"`
		CategoryID string `xml:"categoryId"`
		Category   string `xml:"category"`
		Size       string `xml:"size"`
		Comments   int    `xml:"comments"`
		Trusted    string `xml:"trusted"`
		Remake     string `xml:"remake"`
	}

	// channel represents the channel information containing the items
	channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		Items       []item `xml:"item"`
	}

	// rss is the top-level structure
	rss struct {
		XMLName xml.Name `xml:"rss"`
		Channel channel  `xml:"channel"`
	}
)

// Search implements torrentsource.Source.
// It queries the nyaa RSS feed for the entry titles, then filters and sorts
// the results so the newest/best candidates come first.
func (api *API) Search(
	ctx context.Context,
	entry animelist.Entry,
	opts torrentsource.SearchOptions,
) ([]torrentsource.Torrent, error) {
	var values url.Values

	for name, value := range api.config.CustomParameters {
		values.Set(name, value)
	}

	query := buildQuery(entry, opts)
	values.Add("q", query.String())

	searcher := searcher.New(
		pageSize,
		func(ctx context.Context, offset int) ([]torrentsource.Torrent, error) {
			items, err := api.fetchPage(ctx, values, offset)
			if err != nil {
				return nil, err
			}

			page := sliceutils.Map(items, func(item item) torrentsource.Torrent {
				metadata := parser.Parse(item.Title, 1, opts.Sources)
				publishedAt := must.Must(time.Parse(time.RFC1123Z, item.PubDate))

				return torrentsource.Torrent{
					Title:       item.Title,
					Link:        item.Link,
					Seeders:     item.Seeders,
					PublishedAt: publishedAt,
					Hash:        item.InfoHash,
					Metadata:    metadata,
				}
			})

			return page, nil
		},
		func(ignoreCounter func(string)) func(torrentsource.Torrent) bool {
			return func(torrent torrentsource.Torrent) bool {
				for _, title := range entry.Titles {
					// Remove season information from the original title, as it is not always present in the nyaa entry.
					originalTitleWithoutSeason := parser.StripSeason(title)
					originalTitleWithoutSubtitle := parser.StripSubtitle(
						originalTitleWithoutSeason,
					)

					if stringutils.MatchPrefixFlexible(
						torrent.Metadata.ShowTitle,
						originalTitleWithoutSubtitle,
						torrentsource.IgnoreCharset,
					) {
						return true
					}
				}

				ignoreCounter("titlePrefix")

				return false
			}
		},
	)

	torrents, err := searcher.Search(ctx, entry, opts)
	if err != nil {
		return nil, fmt.Errorf("searching torrent candidates: %w", err)
	}

	return torrents, nil
}

// buildQuery builds the `q` search query from the entry titles,
// qualities, sources and the user suffix, using the nyaaquerier builder.
func buildQuery(entry animelist.Entry, opt torrentsource.SearchOptions) nyaaquerier.And {
	// For title we filter for english and original titles.
	sanitizedTitles := sliceutils.ForEach(entry.Titles,
		strings.ToLower,
		parser.StripTitle,
		parser.StripSubtitle,
		nyaaquerier.Sanitize,
	)

	sort.Strings(sanitizedTitles)
	sanitizedTitles = slices.Compact(sanitizedTitles)

	titleNodes := sliceutils.Map(sanitizedTitles, func(title string) nyaaquerier.Node {
		return nyaaquerier.PhraseOf(title)
	})

	parts := []nyaaquerier.Node{nyaaquerier.Or(titleNodes)}

	if len(opt.Qualities) > 0 {
		qualityNodes := sliceutils.Map(opt.Qualities, func(quality string) nyaaquerier.Node {
			return nyaaquerier.And(sliceutils.Map(strings.Fields(quality), nyaaquerier.PhraseOf))
		})
		parts = append(parts, nyaaquerier.Or(qualityNodes))
	}

	if len(opt.Sources) > 0 {
		sourceNodes := sliceutils.Map(opt.Sources, nyaaquerier.PhraseOf)
		parts = append(parts, nyaaquerier.Or(sourceNodes))
	}

	if opt.SearchSuffix != "" {
		parts = append(parts, nyaaquerier.PhraseOf(opt.SearchSuffix))
	}

	return parts
}

func (api *API) fetchPage(
	ctx context.Context,
	values url.Values,
	offset int,
) ([]item, error) {
	req := must.Must(http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil))

	values.Set("offset", strconv.Itoa(offset))
	values.Set("limit", strconv.Itoa(pageSize))
	req.URL.RawQuery = values.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(must.Must(io.ReadAll(resp.Body))))
	}

	var feed rss
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return feed.Channel.Items, nil
}
