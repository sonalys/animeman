package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type fileRepository struct {
	conn *pgxpool.Pool
}

func NewFileRepository(conn *pgxpool.Pool) ports.FileRepository {
	return &fileRepository{}
}

func (r *fileRepository) Create(ctx context.Context, file *collections.File) error {
	queries := sqlcgen.New(r.conn)

	_, err := queries.RegisterCollectionFile(ctx, sqlcgen.RegisterCollectionFileParams{
		ID:           file.ID,
		EpisodeID:    file.EpisodeID,
		SeasonID:     file.SeasonID,
		MediaID:      file.MediaID,
		RelativePath: file.RelativePath,
		SizeBytes:    file.SizeBytes,
		ReleaseGroup: pgtype.Text{
			String: file.ReleaseGroup,
			Valid:  true,
		},
		Version: int32(file.Version),
		CreatedAt: pgtype.Timestamptz{
			Time:  file.CreatedAt,
			Valid: true,
		},
		Source:          mappers.NewFileSourceModel(file.Source),
		VideoInfo:       mappers.NewVideoInfoModel(file.VideoInfo),
		AudioStreams:    sliceutils.Map(file.AudioStreams, mappers.NewAudioStreamModel),
		SubtitleStreams: sliceutils.Map(file.SubtitleStreams, mappers.NewSubtitleModel),
		Hashes:          sliceutils.Map(file.Hashes, mappers.NewHashModel),
		Chapters:        sliceutils.Map(file.Chapters, mappers.NewChapterModel),
	})
	if err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *fileRepository) Delete(ctx context.Context, id collections.FileID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteCollectionFile(ctx, id); err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *fileRepository) ListByCollection(ctx context.Context, id collections.CollectionID, opts ports.ListOptions) ([]collections.File, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListCollectionFilesPaginated(ctx, sqlcgen.ListCollectionFilesPaginatedParams{
		CollectionID: pgtype.UUID{
			Bytes: id.UUID,
			Valid: true,
		},
		Limit: pgtype.Int4{
			Int32: opts.PageSize.Value(),
			Valid: opts.PageSize.IsSet(),
		},
		LastID: pgtype.UUID{
			Bytes: opts.Cursor.Value().UUID,
			Valid: opts.Cursor.IsSet(),
		},
	})
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.File, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, mappers.NewCollectionFile(model))
	}

	return response, nil
}

func (r *fileRepository) Update(ctx context.Context, id collections.FileID, updateHandler ports.UpdateHandler[collections.File]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetCollectionFile(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		file := mappers.NewCollectionFile(&model)

		if err := updateHandler(&file); err != nil {
			return err
		}

		if _, err := queries.UpdateCollectionFile(ctx, sqlcgen.UpdateCollectionFileParams{
			ID:              id,
			RelativePath:    file.RelativePath,
			SizeBytes:       file.SizeBytes,
			Version:         int32(file.Version),
			VideoInfo:       mappers.NewVideoInfoModel(file.VideoInfo),
			AudioStreams:    sliceutils.Map(file.AudioStreams, mappers.NewAudioStreamModel),
			SubtitleStreams: sliceutils.Map(file.SubtitleStreams, mappers.NewSubtitleModel),
			Chapters:        sliceutils.Map(file.Chapters, mappers.NewChapterModel),
			Hashes:          sliceutils.Map(file.Hashes, mappers.NewHashModel),
		}); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
