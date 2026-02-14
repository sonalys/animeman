package collections

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/hashing"
	"github.com/sonalys/animeman/internal/domain/stream"
)

type (
	FileID struct{ uuid.UUID }

	File struct {
		ID        FileID
		EpisodeID EpisodeID
		SeasonID  SeasonID
		MediaID   MediaID

		RelativePath string
		SizeBytes    int64

		ReleaseGroup string
		Version      uint
		Source       FileSource

		VideoInfo       stream.Video
		AudioStreams    []stream.Audio
		SubtitleStreams []stream.Subtitle
		Hashes          []hashing.Hash

		Chapters []Chapter

		CreatedAt time.Time
	}
)
