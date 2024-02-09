package torrentclient

import (
	"fmt"
)

var ErrUnauthorized = fmt.Errorf("unauthorized")

type (
	Torrent struct {
		Name     string
		Category string
		Hash     string
		Tags     []string
	}

	AddTorrentConfig struct {
		URLs     []string
		Tags     []string
		SavePath string
		Category string
		Paused   bool
	}

	ListTorrentConfig struct {
		Category string
		Tag      string
	}
)
