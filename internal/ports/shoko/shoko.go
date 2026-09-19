package shoko

import (
	"context"
)

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
		// Scanned reports whether shoko already ran its AniDB hash scan on the file,
		// i.e. it found a release info for it. Unlinked + scanned means the hash
		// match failed and the file needs manual linking.
		Scanned bool
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
	// FindSeriesByAnilistID looks up the AniDB id of the shoko series linked to
	// the given AniList id. It returns 0 when nothing matched.
	FindSeriesByAnilistID(ctx context.Context, anilistID int) (anidbID int, err error)
	// FindSeriesByTitle searches the shoko title dump for a series matching the title.
	// It returns the AniDB id of the best match, 0 when nothing matched.
	FindSeriesByTitle(ctx context.Context, title string) (anidbID int, err error)
	// FindEpisodes lists the episode-type episodes of a series by AniDB id.
	FindEpisodes(ctx context.Context, anidbID int) ([]Episode, error)
	// FindFileByPath returns the shoko file whose path ends with the given suffix, nil when shoko has no such file.
	// Linked reports whether the file is already linked to episodes.
	FindFileByPath(ctx context.Context, pathSuffix string) (file *File, linked bool, err error)
	// RescanFile asks shoko to rescan a file, retrying its hash-based AniDB match.
	RescanFile(ctx context.Context, fileID int) error
	// AutoMatchFile asks shoko to run its local filename-based release search
	// on an unrecognized file. It reports whether shoko found a match.
	AutoMatchFile(ctx context.Context, fileID int) (matched bool, err error)
	// LinkFileToEpisodes links an unrecognized file to the given shoko episode ids.
	LinkFileToEpisodes(ctx context.Context, fileID int, episodeIDs []int) error
}
