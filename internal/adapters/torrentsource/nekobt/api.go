package nekobt

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/sonalys/animeman/internal/ports/torrentsource"
	"github.com/sonalys/animeman/internal/utils"
	"github.com/sonalys/animeman/internal/ports/animelist"
)

const TORZNAB_URL = "https://nekobt.to/api/torznab/api"

type Config struct {
	// APIKey is the nekoBT api key, used as the `apikey` query parameter.
	APIKey string
}

type API struct {
	config Config
	client *http.Client
}

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}

// attr is a torznab `<torznab:attr name="..." value="..."/>` element.
type attr struct {
	Name  string `xml:"name,attr"`
	Value string `xml:"value,attr"`
}

// Item represents a single torrent entry in the torznab feed.
type Item struct {
	Title   string `xml:"title"`
	Link    string `xml:"link"`
	GUID    string `xml:"guid"`
	PubDate string `xml:"pubDate"`
	Size    int64  `xml:"size"`
	Attrs   []attr `xml:"attr"`
}

func (i Item) attr(name string) string {
	for _, a := range i.Attrs {
		if a.Name == name {
			return a.Value
		}
	}
	return ""
}

func (i Item) seeders() int {
	v, err := strconv.Atoi(i.attr("seeders"))
	if err != nil {
		return 0
	}
	return v
}

type Channel struct {
	Title string `xml:"title"`
	Items []Item `xml:"item"`
}

type RSS struct {
	XMLName xml.Name `xml:"rss"`
	Channel Channel  `xml:"channel"`
}

// Search implements torrentsource.Source.
// nekoBT matches the entry by AniList/MAL id directly, so no title filtering is needed.
func (api *API) Search(ctx context.Context, entry animelist.Entry, opts torrentsource.SearchOptions) ([]torrentsource.Torrent, error) {
	mediaID, err := resolveMediaID(entry)
	if err != nil {
		return nil, err
	}

	req := utils.Must(http.NewRequestWithContext(ctx, http.MethodGet, TORZNAB_URL, nil))

	q := req.URL.Query()
	q.Set("t", "search")
	q.Set("media_id", mediaID)
	if api.config.APIKey != "" {
		q.Set("apikey", api.config.APIKey)
	}
	if opts.SearchSuffix != "" {
		q.Set("q", opts.SearchSuffix)
	}
	req.URL.RawQuery = q.Encode()

	resp, err := api.client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("fetching response: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("request failed: %s", string(utils.Must(io.ReadAll(resp.Body))))
	}

	var feed RSS
	if err := xml.NewDecoder(resp.Body).Decode(&feed); err != nil {
		return nil, fmt.Errorf("reading response: %w", err)
	}

	return toTorrents(feed.Channel.Items), nil
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

func toTorrents(items []Item) []torrentsource.Torrent {
	return utils.Map(items, func(item Item) torrentsource.Torrent {
		pubDate, err := time.Parse(time.RFC1123Z, item.PubDate)
		if err != nil {
			pubDate = time.Time{}
		}
		return torrentsource.Torrent{
			Title:   item.Title,
			Link:    item.Link,
			PubDate: pubDate,
			Seeders: item.seeders(),
		}
	})
}
