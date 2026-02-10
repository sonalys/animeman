package discovery

import (
	"time"
)

type Config struct {
	Sources          []string
	Qualitites       []string
	Category         string
	RenameTorrent    bool
	DownloadPath     string
	CreateShowFolder bool
	PollFrequency    time.Duration
}
