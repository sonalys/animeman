package postgres

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain"
	"google.golang.org/grpc/codes"
)

type prowlarrRepository struct {
	conn *pgxpool.Pool
}

func prowlarrConfigErrorHandler(err *pgconn.PgError) error {
	switch err.Code {
	case pgerrcode.UniqueViolation:
		switch err.ConstraintName {
		case "prowlarr_configurations_pkey":
			return apperr.New(err, codes.AlreadyExists)
		default:
			return apperr.New(err, codes.FailedPrecondition)
		}
	default:
		return err
	}
}

func (r prowlarrRepository) CreateConfig(ctx context.Context, config *domain.ProwlarrConfiguration) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateProwlarrConfigurationParams{
		ID:      config.ID.String(),
		OwnerID: config.OwnerID.String(),
		Host:    config.Host,
		ApiKey:  config.APIKey,
	}

	if _, err := queries.CreateProwlarrConfiguration(ctx, params); err != nil {
		if err := handleWriteError(err, prowlarrConfigErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r prowlarrRepository) GetConfigByOwner(ctx context.Context, owner domain.UserID) (*domain.ProwlarrConfiguration, error) {
	queries := sqlcgen.New(r.conn)

	prowlarrConfigModel, err := queries.GetProwlarrConfigurationByOwner(ctx, owner.String())
	if err != nil {
		return nil, handleReadError(err)
	}

	return mappers.NewProwlarrConfiguration(&prowlarrConfigModel), nil
}

func (r prowlarrRepository) UpdateConfig(ctx context.Context, id domain.ProwlarrConfigID, update func(config *domain.ProwlarrConfiguration) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	prowlarrConfigModel, err := queries.GetProwlarrConfigurationByOwner(ctx, id.String())
	if err != nil {
		return handleReadError(err)
	}

	prowlarrConfiguration := mappers.NewProwlarrConfiguration(&prowlarrConfigModel)

	if err := update(prowlarrConfiguration); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateProwlarrConfigurationParams{
		ID:     id.String(),
		Host:   prowlarrConfigModel.Host,
		ApiKey: prowlarrConfiguration.APIKey,
	}

	if _, err = queries.UpdateProwlarrConfiguration(ctx, updateParams); err != nil {
		if err := handleWriteError(err, prowlarrConfigErrorHandler); err != nil {
			return err
		}

		return handleReadError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r prowlarrRepository) DeleteConfig(ctx context.Context, id domain.ProwlarrConfigID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteProwlarrConfiguration(ctx, id.String()); err != nil {
		return handleReadError(err)
	}

	return nil
}
