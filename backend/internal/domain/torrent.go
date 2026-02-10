package domain

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
)
