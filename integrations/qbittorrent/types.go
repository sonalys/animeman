package qbittorrent

import (
	"fmt"
	"strings"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

type (
	Torrent struct {
		Name     string `json:"name"`
		Category string `json:"category"`
		Hash     string `json:"hash"`
		Tags     string `json:"tags"`
	}

	TorrentURL []string
	Tag        string
	Tags       []string
	SavePath   string
	Category   string
	Paused     bool
)

func NewErrConnection(err error) error {
	return fmt.Errorf("connection error: %w", err)
}

func (t Torrent) GetTags() []string {
	return strings.Split(t.Tags, ",")
}
