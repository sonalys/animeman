package repositories

import (
	"context"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/optional"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
)

type episodeRepository struct {
	conn *pgxpool.Pool
}

func NewEpisodeRepository(conn *pgxpool.Pool) ports.EpisodeRepository {
	return &episodeRepository{
		conn: conn,
	}
}

func (r *episodeRepository) Create(ctx context.Context, episode *collections.Episode) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateEpisodeParams{
		ID:         episode.ID,
		MediaID:    episode.MediaID,
		SeasonID:   episode.SeasonID,
		Number:     episode.Number,
		Titles:     sliceutils.Map(episode.Titles, mappers.NewTitleModel),
		Type:       mappers.NewMediaTypeModel(episode.Type),
		AiringDate: pgtype.Timestamptz{Time: optional.Coalesce(episode.AiringDate, time.Time{})},
	}

	if _, err := queries.CreateEpisode(ctx, params); err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *episodeRepository) Delete(ctx context.Context, id collections.EpisodeID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteEpisode(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *episodeRepository) ListBySeason(ctx context.Context, id collections.SeasonID) ([]collections.Episode, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListEpisodesBySeason(ctx, id)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Episode, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, collections.Episode{
			ID:         model.ID,
			MediaID:    model.MediaID,
			SeasonID:   model.SeasonID,
			Type:       mappers.NewMediaType(model.Type),
			Number:     model.Number,
			Titles:     sliceutils.Map(model.Titles, mappers.NewTitle),
			AiringDate: model.AiringDate.Time,
		})
	}

	return response, nil
}

func (r *episodeRepository) Update(ctx context.Context, id collections.EpisodeID, updateHandler ports.UpdateHandler[collections.Episode]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetEpisode(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		episode := &collections.Episode{
			ID:         model.ID,
			MediaID:    model.MediaID,
			SeasonID:   model.SeasonID,
			Type:       mappers.NewMediaType(model.Type),
			Number:     model.Number,
			Titles:     sliceutils.Map(model.Titles, mappers.NewTitle),
			AiringDate: model.AiringDate.Time,
		}

		if err := updateHandler(episode); err != nil {
			return err
		}

		params := sqlcgen.UpdateEpisodeMetadataParams{
			ID:         id,
			Type:       mappers.NewMediaTypeModel(episode.Type),
			Titles:     sliceutils.Map(episode.Titles, mappers.NewTitleModel),
			AiringDate: pgtype.Timestamptz{Time: episode.AiringDate},
		}

		if _, err := queries.UpdateEpisodeMetadata(ctx, params); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
