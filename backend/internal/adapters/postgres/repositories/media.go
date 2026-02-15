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

type mediaRepository struct {
	conn *pgxpool.Pool
}

func NewMediaRepository(conn *pgxpool.Pool) ports.MediaRepository {
	return &mediaRepository{
		conn: conn,
	}
}

func (r *mediaRepository) Create(ctx context.Context, media *collections.Media) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateMediaParams{
		ID:               media.ID,
		CollectionID:     media.CollectionID,
		QualityProfileID: media.QualityProfileID,
		Titles:           sliceutils.Map(media.Titles, mappers.NewTitleModel),
		Genres:           media.Metadata.Genres,
		MonitoredSince:   pgtype.Timestamptz{Time: media.MonitoredSince},
		CreatedAt:        pgtype.Timestamptz{Time: media.CreatedAt},
		AiringStartedAt:  pgtype.Timestamptz{Time: media.Metadata.AiringStartedAt},
		AiringEndedAt:    pgtype.Timestamptz{Time: media.Metadata.AiringEndedAt},
		MonitoringStatus: mappers.NewMonitoringStatusModel(media.MonitoringStatus),
	}

	if _, err := queries.CreateMedia(ctx, params); err != nil {
		return handleWriteError(err)
	}

	return nil
}

func (r *mediaRepository) Delete(ctx context.Context, id collections.MediaID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteMedia(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *mediaRepository) ListByCollection(ctx context.Context, id collections.CollectionID, opts ports.ListOptions) ([]collections.Media, error) {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.ListMediaPaginatedParams{
		CollectionID: id,
		LastID:       opts.Cursor,
		Limit:        opts.PageSize,
	}

	models, err := queries.ListMediaPaginated(ctx, params)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Media, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, collections.Media{
			ID:               model.ID,
			CollectionID:     model.CollectionID,
			QualityProfileID: model.QualityProfileID,
			Titles:           sliceutils.Map(model.Titles, mappers.NewTitle),
			MonitoringStatus: mappers.NewMonitoringStatus(model.MonitoringStatus),
			MonitoredSince:   model.MonitoredSince.Time,
			CreatedAt:        model.CreatedAt.Time,
			Metadata: collections.MediaMetadata{
				Genres:          model.Genres,
				AiringStartedAt: model.AiringStartedAt.Time,
				AiringEndedAt:   model.AiringEndedAt.Time,
			},
		})
	}

	return response, nil
}

func (r *mediaRepository) Update(ctx context.Context, id collections.MediaID, updateHandler ports.UpdateHandler[collections.Media]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetMedia(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		media := &collections.Media{
			ID:               model.ID,
			CollectionID:     model.CollectionID,
			QualityProfileID: model.QualityProfileID,
			Titles:           sliceutils.Map(model.Titles, mappers.NewTitle),
			MonitoredSince:   model.MonitoredSince.Time,
			CreatedAt:        model.CreatedAt.Time,
			Metadata: collections.MediaMetadata{
				Genres:          model.Genres,
				AiringStartedAt: model.AiringStartedAt.Time,
				AiringEndedAt:   model.AiringEndedAt.Time,
			},
			MonitoringStatus: mappers.NewMonitoringStatus(model.MonitoringStatus),
		}

		if err := updateHandler(media); err != nil {
			return err
		}

		params := sqlcgen.UpdateMediaParams{
			ID:               id,
			QualityProfileID: media.QualityProfileID,
			Titles:           sliceutils.Map(media.Titles, mappers.NewTitleModel),
			MonitoredSince:   pgtype.Timestamptz{Time: media.MonitoredSince},
			Genres:           media.Metadata.Genres,
			AiringStartedAt:  pgtype.Timestamptz{Time: media.Metadata.AiringStartedAt},
			AiringEndedAt:    pgtype.Timestamptz{Time: media.Metadata.AiringEndedAt},
			MonitoringStatus: mappers.NewMonitoringStatusModel(media.MonitoringStatus),
		}

		if err := queries.UpdateMedia(ctx, params); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
