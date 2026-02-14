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
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"google.golang.org/grpc/codes"
)

type indexerClientRepository struct {
	conn *pgxpool.Pool
}

func indexerClientErrorHandler(err *pgconn.PgError) error {
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

func (r indexerClientRepository) Create(ctx context.Context, config *indexing.IndexerClient) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateProwlarrConfigurationParams{
		ID:      config.ID.String(),
		OwnerID: config.OwnerID.String(),
		Host:    config.Address.String(),
		ApiKey:  config.Authentication.Key,
	}

	if _, err := queries.CreateProwlarrConfiguration(ctx, params); err != nil {
		if err := handleWriteError(err, indexerClientErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r indexerClientRepository) GetByOwner(ctx context.Context, owner shared.UserID) (*indexing.IndexerClient, error) {
	queries := sqlcgen.New(r.conn)

	entityModel, err := queries.GetProwlarrConfigurationByOwner(ctx, owner.String())
	if err != nil {
		return nil, handleReadError(err)
	}

	return mappers.NewIndexerClient(&entityModel), nil
}

func (r indexerClientRepository) Update(ctx context.Context, id indexing.IndexerID, update func(indexerClient *indexing.IndexerClient) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	entityModel, err := queries.GetProwlarrConfigurationByOwner(ctx, id.String())
	if err != nil {
		return handleReadError(err)
	}

	indexerClient := mappers.NewIndexerClient(&entityModel)

	if err := update(indexerClient); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateProwlarrConfigurationParams{
		ID:     id.String(),
		Host:   entityModel.Host,
		ApiKey: indexerClient.Authentication.Key,
	}

	if _, err = queries.UpdateProwlarrConfiguration(ctx, updateParams); err != nil {
		if err := handleWriteError(err, indexerClientErrorHandler); err != nil {
			return err
		}

		return handleReadError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}

func (r indexerClientRepository) Delete(ctx context.Context, id indexing.IndexerID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteProwlarrConfiguration(ctx, id.String()); err != nil {
		return handleReadError(err)
	}

	return nil
}
