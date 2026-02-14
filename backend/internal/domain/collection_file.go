package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	CollectionFileID struct{ uuid.UUID }

	CollectionFile struct {
		ID        CollectionFileID
		EpisodeID EpisodeID
		SeasonID  SeasonID
		MediaID   MediaID

		RelativePath string
		SizeBytes    int64

		ReleaseGroup string
		Version      uint
		Source       FileSource

		Video           VideoInfo
		AudioStreams    []AudioStream
		SubtitleStreams []SubtitleStream
		Hashes          []Hash

		Chapters []Chapter

		CreatedAt time.Time
	}
)
