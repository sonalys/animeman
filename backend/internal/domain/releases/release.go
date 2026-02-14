package releases

import (
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/hashing"
	"github.com/sonalys/animeman/internal/domain/stream"
)

type (
	ReleaseID struct{ uuid.UUID }

	Release struct {
		ID      ReleaseID
		MediaID collections.MediaID

		Title     string
		Hash      hashing.Hash
		SizeBytes int64
		URL       url.URL
		Metadata  Metadata
	}

	Metadata struct {
		VideoStream     stream.Video
		AudioStreams    []stream.Audio
		SubtitleStreams []stream.Subtitle
	}
)
