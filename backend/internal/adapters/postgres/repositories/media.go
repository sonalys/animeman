package repositories

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
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
		if err := handleWriteError(err, mediaErrorHandler); err != nil {
			return err
		}

		return err
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

	entityModels, err := queries.ListMediaPaginated(ctx, sqlcgen.ListMediaPaginatedParams{
		LastID: opts.Cursor,
		Limit:  opts.PageSize,
	})
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Media, 0, len(entityModels))

	for i := range entityModels {
		item := entityModels[i]
		response = append(response, collections.Media{
			ID:               item.ID,
			CollectionID:     item.CollectionID,
			QualityProfileID: item.QualityProfileID,
			Titles:           sliceutils.Map(item.Titles, mappers.NewTitle),
			MonitoringStatus: mappers.NewMonitoringStatus(item.MonitoringStatus),
			MonitoredSince:   item.MonitoredSince.Time,
			CreatedAt:        item.CreatedAt.Time,
			Metadata: collections.MediaMetadata{
				Genres:          item.Genres,
				AiringStartedAt: item.AiringStartedAt.Time,
				AiringEndedAt:   item.AiringEndedAt.Time,
			},
		})
	}

	return response, nil
}

func (r *mediaRepository) Update(ctx context.Context, id collections.MediaID, updateHandler ports.UpdateHandler[collections.Media]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		mediaModel, err := queries.GetMedia(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		media := &collections.Media{
			ID:               mediaModel.ID,
			CollectionID:     mediaModel.CollectionID,
			QualityProfileID: mediaModel.QualityProfileID,
			Titles:           sliceutils.Map(mediaModel.Titles, mappers.NewTitle),
			MonitoredSince:   mediaModel.MonitoredSince.Time,
			CreatedAt:        mediaModel.CreatedAt.Time,
			Metadata: collections.MediaMetadata{
				Genres:          mediaModel.Genres,
				AiringStartedAt: mediaModel.AiringStartedAt.Time,
				AiringEndedAt:   mediaModel.AiringEndedAt.Time,
			},
			MonitoringStatus: mappers.NewMonitoringStatus(mediaModel.MonitoringStatus),
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
			if err := handleWriteError(err, mediaErrorHandler); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}

func mediaErrorHandler(err *pgconn.PgError) error {
	switch err.Code {
	case pgerrcode.UniqueViolation:
		switch err.ConstraintName {
		default:
			return apperr.New(err, codes.FailedPrecondition)
		}
	default:
		return err
	}
}
