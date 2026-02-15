package collections

import (
	"time"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/hashing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/stream"
)

type (
	EpisodeID struct{ uuid.UUID }

	Episode struct {
		ID       EpisodeID
		SeasonID SeasonID
		MediaID  MediaID
		Type     MediaType

		// Number is a string due to complex episode number variations, like episode 6.5 or combined episodes.
		Number     string
		Titles     []Title
		AiringDate time.Time
	}
)

func (e *Episode) NewFile(
	relativePath string,
	sizeBytes int64,
	releaseGroup string,
	version uint,
	source FileSource,
	videoInfo stream.Video,
	audioStreams []stream.Audio,
	subtitleStreams []stream.Subtitle,
	hashes []hashing.Hash,
	chapters []Chapter,
) *File {
	file := &File{
		ID:              shared.NewID[FileID](),
		MediaID:         e.MediaID,
		SeasonID:        e.SeasonID,
		EpisodeID:       e.ID,
		RelativePath:    relativePath,
		SizeBytes:       sizeBytes,
		ReleaseGroup:    releaseGroup,
		Version:         version,
		Source:          source,
		VideoInfo:       videoInfo,
		AudioStreams:    audioStreams,
		SubtitleStreams: subtitleStreams,
		Hashes:          hashes,
		Chapters:        chapters,
		CreatedAt:       time.Now(),
	}

	return file
}
