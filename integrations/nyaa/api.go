package nyaa

import (
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

func New(client *http.Client, c Config) *API {
	return &API{
		config: c,
		client: client,
	}
}
