package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	FileID struct{ uuid.UUID }

	LocalFile struct {
		ID FileID

		Path string
		Size int64
		Hash string

		ReleaseGroup string
		Quality      QualityInfo

		CreatedAt time.Time
	}
)

func (e *Episode) BestFile() *LocalFile {
	if len(e.Files) == 0 {
		return nil
	}
	best := &e.Files[0]
	for i := 1; i < len(e.Files); i++ {
		if e.Files[i].Quality.Resolution > best.Quality.Resolution {
			best = &e.Files[i]
			continue
		}

		if e.Files[i].Quality.Codec > best.Quality.Codec {
			best = &e.Files[i]
			continue
		}
	}
	return best
}
