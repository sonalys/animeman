package shoko

import (
	"context"
	"errors"
	"time"
)

var ErrForbidden = errors.New("forbidden")

type (
	// Episode is an AniDB episode known by shoko.
	Episode struct {
		// AniDBID is the AniDB episode id.
		AniDBID int
		// ShokoEpisodeID is the shoko internal episode id, used for linking files.
		ShokoEpisodeID int
		// Number is the AniDB episode number.
		Number int
	}

	// File is a video file known by shoko.
	File struct {
		// ID is the shoko internal file id.
		ID int
		// RelativePath is the path of the file relative to its managed folder.
		RelativePath string
		CreatedAt    time.Time
	}

	// Config controls how the shoko adapter connects to the server.
	Config struct {
		// Host is the base url of the shoko server.
		Host string
		// APIKey is the shoko api key, sent as the `apikey` header.
		APIKey string
	}
)

// Shoko is the port implemented by the shoko server adapter.
type Shoko interface {
	// Wait blocks until shoko is reachable or the context is cancelled.
	Wait(ctx context.Context)
	// AutoMatchFile asks shoko to run its local filename-based release search
	// on an unrecognized file. It reports whether shoko found a match.
	AutoMatchFile(ctx context.Context, fileID int) (matched bool, err error)
	ListUnknownFiles(ctx context.Context) ([]File, error)
}
