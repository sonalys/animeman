package nyaa

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"slices"
	"sort"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
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
	items, err := api.list(ctx, entry, opts)
	if err != nil {
		return nil, fmt.Errorf("listing nyaa: %w", err)
	}

	items = utils.Filter(items,
		filterSeeders(1),
		filterMetadata(entry),
	)

	torrents := utils.Map(items, func(item item) torrentsource.Torrent {
		pubDate := utils.Must(time.Parse(time.RFC1123Z, item.PubDate))
		return torrentsource.Torrent{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: pubDate,
			Seeders: item.Seeders,
		}
	})

	torrents = parser.Prioritize(entry, torrents, opts)

	return torrents, nil
}

func buildQuery(entry animelist.Entry, opt torrentsource.SearchOptions) string {
	var b strings.Builder

	titleSanitization := strings.NewReplacer(
		"-", " ",
		"\"", " ",
		"'", " ",
		"(", " ",
		")", " ",
	)

	// Build search query for Nyaa.
	// For title we filter for english and original titles.
	sanitizedTitles := utils.Transform(entry.Titles,
		strings.ToLower,
		parser.StripTitle,
		parser.StripSubtitle,
		titleSanitization.Replace,
	)

	sort.Strings(sanitizedTitles)
	sanitizedTitles = slices.Compact(sanitizedTitles)

	titles := utils.Map(sanitizedTitles, func(from string) string { return "(" + from + ")" })
	fmt.Fprintf(&b, "%s", strings.Join(titles, "|"))

	if resolutions := opt.Qualities; len(resolutions) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(resolutions, "|"))
	}

	if sources := opt.Sources; len(sources) > 0 {
		fmt.Fprintf(&b, " (%s)", strings.Join(sources, "|"))
	}

	if opt.SearchSuffix != "" {
		fmt.Fprintf(&b, " %s", opt.SearchSuffix)
	}

	return b.String()
}

func (api *API) list(
	ctx context.Context,
	entry animelist.Entry,
	options torrentsource.SearchOptions,
) ([]item, error) {
	var path = API_URL

	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, path, nil))

	q := req.URL.Query()
	for name, value := range api.config.ListParameters {
		q.Set(name, value)
	}

	q.Add("q", buildQuery(entry, options))

	req.URL.RawQuery = q.Encode()

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

	return feed.Channel.Items, nil
}

func filterSeeders(minSeeders int) func(item) bool {
	return func(item item) bool {
		return item.Seeders >= minSeeders
	}
}

// filterMetadata ensures that only coherent and expected nyaa entries are considered for download.
// This function avoids downloading unrelated torrents.
func filterMetadata(
	entry animelist.Entry,
) func(e item) bool {
	return func(nyaaEntry item) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, nyaaEntry.PubDate))

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if !publishedDate.IsZero() && publishedDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		// Check if nyaa entry episode is greater than the animelist episode count.
		if entry.NumEpisodes > 0 &&
			parser.Parse(nyaaEntry.Title, 1, nil).Tag.LastEpisode() > float64(entry.NumEpisodes) {
			return false
		}

		nyaaTitleWithoutTags := parser.StripTags(nyaaEntry.Title)

		for _, originalTitle := range entry.Titles {
			// Remove season information from the original title, as it is not always present in the nyaa entry.
			originalTitleWithoutSeason := parser.StripSeason(originalTitle)
			originalTitleWithoutSubtitle := parser.StripSubtitle(originalTitleWithoutSeason)

			if utils.MatchPrefixFlexible(
				nyaaTitleWithoutTags,
				originalTitleWithoutSubtitle,
				parser.IgnoreCharset,
			) {
				return true
			}
		}

		return false
	}
}
