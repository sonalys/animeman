package domain

import "github.com/gofrs/uuid/v5"

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
		Name     *string
		SavePath string
		Category string
		Paused   bool
	}

	ListTorrentConfig struct {
		Category *string
		Tag      *string
	}

	TorrentClientID struct{ uuid.UUID }
	TorrentSource   uint

	TorrentClient struct {
		ID       TorrentClientID
		OwnerID  UserID
		Source   TorrentSource
		Host     string
		Username string
		Password []byte
	}
)
