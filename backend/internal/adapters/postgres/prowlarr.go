package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain"
	"google.golang.org/grpc/codes"
)

type prowlarrRepository struct {
	conn *pgxpool.Pool
}

func (r prowlarrRepository) CreateConfig(ctx context.Context, config *domain.ProwlarrConfiguration) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateProwlarrConfigurationParams{
		ID:      config.ID.String(),
		OwnerID: config.OwnerID.String(),
		Host:    config.Host,
		ApiKey:  config.APIKey,
	}

	_, err := queries.CreateProwlarrConfiguration(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "fk_prowlarr_configuration_owner":
				return apperr.New(err, codes.FailedPrecondition, "owner does not exist")
			}
		}
		return apperr.New(err, codes.Internal, "could not create prowlarr configuration")
	}

	return nil
}

func (r prowlarrRepository) GetConfigByOwner(ctx context.Context, owner domain.UserID) (*domain.ProwlarrConfiguration, error) {
	queries := sqlcgen.New(r.conn)

	prowlarrConfigModel, err := queries.GetProwlarrConfigurationByOwner(ctx, owner.String())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, apperr.New(err, codes.NotFound, "not found")
		default:
			return nil, apperr.New(err, codes.Internal, "could not get prowlarr configuration")
		}
	}

	prowlarrConfiguration := &domain.ProwlarrConfiguration{
		ID:      domain.ProwlarrConfigID{uuid.FromStringOrNil(prowlarrConfigModel.ID)},
		OwnerID: domain.UserID{uuid.FromStringOrNil(prowlarrConfigModel.OwnerID)},
		Host:    prowlarrConfigModel.Host,
		APIKey:  prowlarrConfigModel.ApiKey,
	}

	return prowlarrConfiguration, nil
}

func (r prowlarrRepository) UpdateConfig(ctx context.Context, id domain.ProwlarrConfigID, update func(config *domain.ProwlarrConfiguration) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return apperr.New(err, codes.Internal, "could not start transaction")
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	prowlarrConfigModel, err := queries.GetProwlarrConfigurationByOwner(ctx, id.String())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "not found")
		default:
			return apperr.New(err, codes.Internal, "could not get prowlarr configuration")
		}
	}

	prowlarrConfiguration := &domain.ProwlarrConfiguration{
		ID:      domain.ProwlarrConfigID{uuid.FromStringOrNil(prowlarrConfigModel.ID)},
		OwnerID: domain.UserID{uuid.FromStringOrNil(prowlarrConfigModel.OwnerID)},
		Host:    prowlarrConfigModel.Host,
		APIKey:  prowlarrConfigModel.ApiKey,
	}

	if err := update(prowlarrConfiguration); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateProwlarrConfigurationParams{
		ID:     id.String(),
		Host:   prowlarrConfigModel.Host,
		ApiKey: prowlarrConfiguration.APIKey,
	}

	if _, err = queries.UpdateProwlarrConfiguration(ctx, updateParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.New(err, codes.NotFound, "not found")
		}

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "fk_prowlarr_configuration_owner":
				return apperr.New(err, codes.FailedPrecondition, "owner does not exist")
			}
		}
		return apperr.New(err, codes.Internal, "could not create prowlarr configuration")
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.New(err, codes.Internal, "could not commit transaction")
	}

	return nil
}

func (r prowlarrRepository) DeleteConfig(ctx context.Context, id domain.ProwlarrConfigID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteProwlarrConfiguration(ctx, id.String()); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "not found")
		default:
			return apperr.New(err, codes.Internal, "could not get prowlarr configuration")
		}
	}

	return nil
}
