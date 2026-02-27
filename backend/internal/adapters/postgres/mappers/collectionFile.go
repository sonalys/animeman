package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/dtos"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/hashing"
	"github.com/sonalys/animeman/internal/domain/stream"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

func NewCollectionFile(from *sqlcgen.CollectionFile) collections.File {
	return collections.File{
		ID:           from.ID,
		EpisodeID:    from.EpisodeID,
		SeasonID:     from.SeasonID,
		MediaID:      from.MediaID,
		RelativePath: from.RelativePath,
		SizeBytes:    from.SizeBytes,
		ReleaseGroup: from.ReleaseGroup.String,
		Version:      uint(from.Version),
		Source: func() collections.FileSource {
			switch from.Source {
			case sqlcgen.FileSourceBr:
				return collections.FileSourceBluRay
			case sqlcgen.FileSourceDvd:
				return collections.FileSourceDVD
			case sqlcgen.FileSourceTv:
				return collections.FileSourceTV
			case sqlcgen.FileSourceWeb:
				return collections.FileSourceWEB
			default:
				return collections.FileSourceUnknown
			}
		}(),
		VideoInfo: stream.Video{
			Codec: func() stream.VideoCodec {
				switch from.VideoInfo.Codec {
				case stream.VideoCodecX264.String():
					return stream.VideoCodecX264
				case stream.VideoCodecX265.String():
					return stream.VideoCodecX265
				case stream.VideoCodecAV1.String():
					return stream.VideoCodecAV1
				default:
					return stream.VideoCodecUnknown
				}
			}(),
		},
		AudioStreams: sliceutils.Map(from.AudioStreams, func(from dtos.AudioStream) stream.Audio {
			return stream.Audio{
				Language: from.Language,
				Title:    from.Title,
				Codec: func() stream.AudioCodec {
					switch from.Codec {
					case stream.AudioCodecAAC.String():
						return stream.AudioCodecAAC
					case stream.AudioCodecAC3.String():
						return stream.AudioCodecAC3
					case stream.AudioCodecDTS.String():
						return stream.AudioCodecDTS
					case stream.AudioCodecFLAC.String():
						return stream.AudioCodecFLAC
					case stream.AudioCodecMP3.String():
						return stream.AudioCodecMP3
					case stream.AudioCodecOpus.String():
						return stream.AudioCodecOpus
					case stream.AudioCodecTrueHD.String():
						return stream.AudioCodecTrueHD
					default:
						return stream.AudioCodecUnknown
					}
				}(),
				Channels: from.Channels,
				BitRate:  from.BitRate,
			}
		}),
		SubtitleStreams: sliceutils.Map(from.SubtitleStreams, func(from dtos.Subtitle) stream.Subtitle {
			return stream.Subtitle{
				Language: from.Language,
				Title:    from.Title,
				Format: func() stream.SubtitleFormat {
					switch from.Format {
					case stream.SubtitleFormatSRT.String():
						return stream.SubtitleFormatSRT
					case stream.SubtitleFormatASS.String():
						return stream.SubtitleFormatASS
					case stream.SubtitleFormatSSA.String():
						return stream.SubtitleFormatSSA
					case stream.SubtitleFormatPGS.String():
						return stream.SubtitleFormatPGS
					case stream.SubtitleFormatVobSub.String():
						return stream.SubtitleFormatVobSub
					default:
						return stream.SubtitleFormatUnknown
					}
				}(),
			}
		}),
		Hashes: sliceutils.Map(from.Hashes, func(from dtos.Hash) hashing.Hash {
			return hashing.Hash{
				Algorithm: func() hashing.HashAlgorithm {
					switch from.Algorithm {
					case hashing.HashAlgCRC32.String():
						return hashing.HashAlgCRC32
					case hashing.HashAlgMD5.String():
						return hashing.HashAlgMD5
					case hashing.HashAlgSHA1.String():
						return hashing.HashAlgSHA1
					case hashing.HashAlgED2K.String():
						return hashing.HashAlgED2K
					case hashing.HashAlgSHA256.String():
						return hashing.HashAlgSHA256
					default:
						return hashing.HashAlgUnknown
					}
				}(),
				Value: from.Value,
			}
		}),
		Chapters: sliceutils.Map(from.Chapters, func(from dtos.Chapter) collections.Chapter {
			return collections.Chapter{
				Title:     from.Title,
				StartTime: from.StartTime,
				EndTime:   from.EndTime,
			}
		}),
		CreatedAt: from.CreatedAt.Time,
	}
}

func NewFileSourceModel(from collections.FileSource) sqlcgen.FileSource {
	switch from {
	case collections.FileSourceBluRay:
		return sqlcgen.FileSourceBr
	case collections.FileSourceDVD:
		return sqlcgen.FileSourceDvd
	case collections.FileSourceTV:
		return sqlcgen.FileSourceTv
	case collections.FileSourceWEB:
		return sqlcgen.FileSourceWeb
	default:
		return sqlcgen.FileSourceUnknown
	}
}

func NewVideoInfoModel(from stream.Video) dtos.VideoInfo {
	return dtos.VideoInfo{
		Codec:      from.Codec.String(),
		Resolution: from.Resolution.String(),
		BitDepth:   from.BitDepth,
		BitRate:    from.BitRate,
		Width:      from.Width,
		Height:     from.Height,
	}
}

func NewAudioStreamModel(from stream.Audio) dtos.AudioStream {
	return dtos.AudioStream{
		Language: from.Language,
		Title:    from.Title,
		Codec:    from.Codec.String(),
		Channels: from.Channels,
		BitRate:  from.BitRate,
	}
}

func NewSubtitleModel(from stream.Subtitle) dtos.Subtitle {
	return dtos.Subtitle{
		Language: from.Language,
		Title:    from.Title,
		Format:   from.Format.String(),
	}
}

func NewHashModel(from hashing.Hash) dtos.Hash {
	return dtos.Hash{
		Algorithm: from.Algorithm.String(),
		Value:     from.Value,
	}
}

func NewChapterModel(from collections.Chapter) dtos.Chapter {
	return dtos.Chapter{
		Title:     from.Title,
		StartTime: from.StartTime,
		EndTime:   from.EndTime,
	}
}
