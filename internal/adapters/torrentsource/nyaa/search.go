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

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/parser"
	"github.com/sonalys/animeman/internal/ports/animelist"
	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/tags"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/utils/nyaaquerier"
	"github.com/sonalys/animeman/internal/utils/pager"
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

	torrents, err := pager.Search(ctx, pager.Page[item]{
		PageSize: pageSize,
		Fetch: func(ctx context.Context, offset int) ([]item, error) {
			return api.fetchPage(ctx, values, offset)
		},
		Filter: func(item item) bool {
			return filterSeeders(1)(item) &&
				filterMetadata(entry, opts.Sources)(item) &&
				filterSources(opts.Sources)(item)
		},
		Map: func(item item) torrentsource.Torrent {
			metadata := parser.Parse(item.Title, 1, opts.Sources)

			return torrentsource.Torrent{
				Title:    item.Title,
				Link:     item.Link,
				Seeders:  item.Seeders,
				Hash:     item.InfoHash,
				Metadata: metadata,
			}
		},
		ShouldPaginate: func(filtered []item) bool {
			return shouldPaginate(filtered, opts.LatestTag)
		},
	})
	if err != nil {
		return nil, err
	}

	torrents = torrentsource.Prioritize(entry, torrents, opts)

	log.
		Ctx(ctx).
		Debug().
		Int("results", len(torrents)).
		Msg("search results")

	return torrents, nil
}

// buildQuery builds the `q` search query from the entry titles,
// qualities, sources and the user suffix, using the nyaaquerier builder.
func buildQuery(entry animelist.Entry, opt torrentsource.SearchOptions) nyaaquerier.And {
	// For title we filter for english and original titles.
	sanitizedTitles := utils.Transform(entry.Titles,
		strings.ToLower,
		parser.StripTitle,
		parser.StripSubtitle,
		nyaaquerier.Sanitize,
	)

	sort.Strings(sanitizedTitles)
	sanitizedTitles = slices.Compact(sanitizedTitles)

	titleNodes := utils.Map(sanitizedTitles, func(title string) nyaaquerier.Node {
		return nyaaquerier.PhraseOf(title)
	})

	parts := []nyaaquerier.Node{nyaaquerier.Or(titleNodes)}

	if len(opt.Qualities) > 0 {
		qualityNodes := utils.Map(opt.Qualities, func(quality string) nyaaquerier.Node {
			return nyaaquerier.And(utils.Map(strings.Fields(quality), nyaaquerier.PhraseOf))
		})
		parts = append(parts, nyaaquerier.Or(qualityNodes))
	}

	if len(opt.Sources) > 0 {
		sourceNodes := utils.Map(opt.Sources, nyaaquerier.PhraseOf)
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
	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, API_URL, nil))

	values.Set("offset", strconv.Itoa(offset))
	values.Set("limit", strconv.Itoa(pageSize))
	req.URL.RawQuery = values.Encode()

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
	sources []string,
) func(e item) bool {
	return func(item item) bool {
		publishedDate := utils.Must(time.Parse(time.RFC1123Z, item.PubDate))

		// Compares publishing date with anime start date, 2 days offset to prevent wrong timezone and hour precision.
		if !publishedDate.IsZero() && publishedDate.Before(entry.StartDate.AddDate(0, 0, -2)) {
			return false
		}

		// If ep number is greater than season ep count, should be removed.
		// This can happen when certain sources mark S2 but use absolute ep number, so they start like S2E13 instead of S2E01.
		// If there's only a single source, then this won't be a problem.
		if len(sources) > 1 && entry.NumEpisodes != 0 {
			metadata := parser.Parse(item.Title, 1, sources)
			if metadata.Tag.FirstEpisode() > float64(entry.NumEpisodes) {
				return false
			}
		}

		nyaaTitleWithoutTags := parser.StripTags(item.Title)

		for _, originalTitle := range entry.Titles {
			// Remove season information from the original title, as it is not always present in the nyaa entry.
			originalTitleWithoutSeason := parser.StripSeason(originalTitle)
			originalTitleWithoutSubtitle := parser.StripSubtitle(originalTitleWithoutSeason)

			if utils.MatchPrefixFlexible(
				nyaaTitleWithoutTags,
				originalTitleWithoutSubtitle,
				torrentsource.IgnoreCharset,
			) {
				return true
			}
		}

		return false
	}
}

// filterSources keeps only torrents whose title contains one of the
// configured sources (release groups), mirroring how the parser extracts
// the release group.
func filterSources(sources []string) func(item) bool {
	return func(item item) bool {
		if len(sources) == 0 {
			return true
		}
		entry := parser.Parse(item.Title, 1, sources)
		return entry.ReleaseGroup != "" && slices.Contains(sources, entry.ReleaseGroup)
	}
}

// shouldPaginate reports whether the source may have more results after this
// page. nekoBT returns newest results first, so paginate while even the
// smallest tag found remains newer than the latest downloaded tag.
func shouldPaginate(items []item, latestTag tags.Tag) bool {
	if latestTag.IsZero() {
		// Nothing downloaded yet: the first page already has everything.
		return false
	}

	if len(items) == 0 {
		return false
	}

	var smallest tags.Tag

	for _, it := range items {
		tag := parser.Parse(it.Title, 1, nil).Tag
		if smallest.IsZero() || tag.Compare(smallest) < 0 {
			smallest = tag
		}
	}

	return smallest.Compare(latestTag) > 0
}
