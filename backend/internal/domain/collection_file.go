package domain

import (
	"time"

	"github.com/gofrs/uuid/v5"
)

type (
	CollectionFileID struct{ uuid.UUID }

	VideoInfo struct {
		Codec      VideoCodec
		Resolution Resolution
		BitDepth   uint
		BitRate    uint
		Width      uint
		Height     uint
	}

	FileHashes struct {
		ED2K string
		CRC  string
		SHA1 string
		MD5  string
	}

	CollectionFile struct {
		ID        CollectionFileID
		EpisodeID EpisodeID
		SeasonID  SeasonID
		MediaID   MediaID

		RelativePath string
		SizeBytes    int64

		Quality      QualityInfo
		ReleaseGroup string
		Version      uint

		Video           VideoInfo
		AudioStreams    []AudioStream
		SubtitleStreams []SubtitleStream
		Hashes          FileHashes

		Chapters []Chapter

		CreatedAt time.Time
	}
)
