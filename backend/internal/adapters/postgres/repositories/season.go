package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/dtos"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
)

type seasonRepository struct {
	conn *pgxpool.Pool
}

func NewSeasonRepository(conn *pgxpool.Pool) ports.SeasonRepository {
	return &seasonRepository{
		conn: conn,
	}
}

func (r *seasonRepository) Create(ctx context.Context, season *collections.Season) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateSeasonParams{
		ID:           season.ID,
		MediaID:      season.MediaID,
		Number:       int32(season.Number),
		AiringStatus: mappers.NewAiringStatusModel(season.AiringStatus),
		Metadata: dtos.SeasonMetadata{
			Season: season.SeasonMetadata.Season,
			Year:   season.SeasonMetadata.Year,
			Month:  season.SeasonMetadata.Month,
			Day:    season.SeasonMetadata.Day,
		},
	}

	if _, err := queries.CreateSeason(ctx, params); err != nil {
		if err := handleWriteError(err); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r *seasonRepository) Delete(ctx context.Context, id collections.SeasonID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteSeason(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *seasonRepository) ListByMedia(ctx context.Context, id collections.MediaID) ([]collections.Season, error) {
	queries := sqlcgen.New(r.conn)

	entityModels, err := queries.ListSeasonsByMedia(ctx, id)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Season, 0, len(entityModels))

	for i := range entityModels {
		item := entityModels[i]
		response = append(response, collections.Season{
			ID:           item.ID,
			MediaID:      item.MediaID,
			Number:       int(item.Number),
			AiringStatus: mappers.NewAiringStatus(item.AiringStatus),
			SeasonMetadata: collections.SeasonMetadata{
				Season: item.Metadata.Season,
				Year:   item.Metadata.Year,
				Month:  item.Metadata.Month,
				Day:    item.Metadata.Day,
			},
		})
	}

	return response, nil
}

func (r *seasonRepository) Update(ctx context.Context, id collections.SeasonID, updateHandler ports.UpdateHandler[collections.Season]) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		seasonModel, err := queries.GetSeason(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		season := &collections.Season{
			ID:           seasonModel.ID,
			MediaID:      seasonModel.MediaID,
			Number:       int(seasonModel.Number),
			AiringStatus: mappers.NewAiringStatus(seasonModel.AiringStatus),
			SeasonMetadata: collections.SeasonMetadata{
				Season: seasonModel.Metadata.Season,
				Year:   seasonModel.Metadata.Year,
				Month:  seasonModel.Metadata.Month,
				Day:    seasonModel.Metadata.Day,
			},
		}

		if err := updateHandler(season); err != nil {
			return err
		}

		params := sqlcgen.UpdateSeasonMetadataParams{
			ID:           id,
			AiringStatus: mappers.NewAiringStatusModel(season.AiringStatus),
			Metadata: dtos.SeasonMetadata{
				Season: season.SeasonMetadata.Season,
				Year:   season.SeasonMetadata.Year,
				Month:  season.SeasonMetadata.Month,
				Day:    season.SeasonMetadata.Day,
			},
		}

		if _, err := queries.UpdateSeasonMetadata(ctx, params); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
