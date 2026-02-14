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
		AiringDate *time.Time
		Files      []*File
	}
)

func (e *Episode) BestFile() *File {
	if len(e.Files) == 0 {
		return nil
	}
	best := e.Files[0]
	for i := 1; i < len(e.Files); i++ {
		if e.Files[i].VideoInfo.Resolution > best.VideoInfo.Resolution {
			best = e.Files[i]
			continue
		}

		if e.Files[i].VideoInfo.Codec > best.VideoInfo.Codec {
			best = e.Files[i]
			continue
		}
	}
	return best
}

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
	e.Files = append(e.Files, file)

	return file
}
