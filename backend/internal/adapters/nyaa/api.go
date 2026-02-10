package nyaa

import (
	"encoding/xml"
	"net/http"
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

	// Item represents a single torrent entry in the RSS feed
	Item struct {
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

	// Channel represents the channel information containing the items
	Channel struct {
		Title       string `xml:"title"`
		Description string `xml:"description"`
		Link        string `xml:"link"`
		Items       []Item `xml:"item"`
	}

	// RSS is the top-level structure
	RSS struct {
		XMLName xml.Name `xml:"rss"`
		Channel Channel  `xml:"channel"`
	}
)

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}
