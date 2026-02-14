package repositories

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/sliceutils"
	"google.golang.org/grpc/codes"
)

type qualityProfileRepository struct {
	conn *pgxpool.Pool
}

func NewQualityProfileRepository(conn *pgxpool.Pool) ports.QualityProfileRepository {
	return &qualityProfileRepository{
		conn: conn,
	}
}

func (r *qualityProfileRepository) Create(ctx context.Context, profile *collections.QualityProfile) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateQualityProfileParams{
		ID:                     profile.ID,
		Name:                   profile.Name,
		MinResolution:          mappers.NewResolutionModel(profile.MinResolution),
		MaxResolution:          mappers.NewResolutionModel(profile.MaxResolution),
		CodecPreference:        sliceutils.Map(profile.CodecPreference, mappers.NewVideoCodecModel),
		ReleaseGroupPreference: profile.ReleaseGroupPreference,
	}

	if _, err := queries.CreateQualityProfile(ctx, params); err != nil {
		if err := handleWriteError(err, qualityProfileErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r *qualityProfileRepository) Delete(ctx context.Context, id collections.QualityProfileID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteQualityProfile(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *qualityProfileRepository) List(ctx context.Context) ([]collections.QualityProfile, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListQualityProfiles(ctx)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.QualityProfile, 0, len(models))

	for i := range models {
		item := models[i]

		response = append(response, collections.QualityProfile{
			ID:                     item.ID,
			Name:                   item.Name,
			MinResolution:          mappers.NewResolution(item.MinResolution),
			MaxResolution:          mappers.NewResolution(item.MaxResolution),
			CodecPreference:        sliceutils.Map(item.CodecPreference, mappers.NewVideoCodec),
			ReleaseGroupPreference: item.ReleaseGroupPreference,
		})
	}

	return response, nil
}

func (r *qualityProfileRepository) Update(ctx context.Context, id collections.QualityProfileID, update func(profile *collections.QualityProfile) error) error {
	return transaction(ctx, r.conn, func(tx pgx.Tx) error {
		queries := sqlcgen.New(tx)

		model, err := queries.GetQualityProfile(ctx, id)
		if err != nil {
			return err
		}

		qualityProfile := collections.QualityProfile{
			ID:                     model.ID,
			Name:                   model.Name,
			MinResolution:          mappers.NewResolution(model.MinResolution),
			MaxResolution:          mappers.NewResolution(model.MaxResolution),
			CodecPreference:        sliceutils.Map(model.CodecPreference, mappers.NewVideoCodec),
			ReleaseGroupPreference: model.ReleaseGroupPreference,
		}

		if err := update(&qualityProfile); err != nil {
			return err
		}

		params := sqlcgen.UpdateQualityProfileParams{
			ID:                     id,
			Name:                   qualityProfile.Name,
			MinResolution:          mappers.NewResolutionModel(qualityProfile.MinResolution),
			MaxResolution:          mappers.NewResolutionModel(qualityProfile.MaxResolution),
			CodecPreference:        sliceutils.Map(qualityProfile.CodecPreference, mappers.NewVideoCodecModel),
			ReleaseGroupPreference: qualityProfile.ReleaseGroupPreference,
		}

		if _, err := queries.UpdateQualityProfile(ctx, params); err != nil {
			return err
		}

		return nil
	})
}

func qualityProfileErrorHandler(err *pgconn.PgError) error {
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
