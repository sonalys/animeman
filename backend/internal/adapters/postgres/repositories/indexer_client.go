package repositories

import (
	"context"
	"net/url"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/errutils"
	"google.golang.org/grpc/codes"
)

type indexerClientRepository struct {
	conn *pgxpool.Pool
}

func NewIndexerClientRepository(conn *pgxpool.Pool) ports.IndexerClientRepository {
	return &indexerClientRepository{
		conn: conn,
	}
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

func (r indexerClientRepository) Create(ctx context.Context, client *indexing.IndexerClient) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateIndexerClientParams{
		ID:      client.ID.String(),
		OwnerID: client.OwnerID.String(),
		Address: client.Address.String(),
		Type:    sqlcgen.IndexerClientType(client.Type.String()),
	}

	if _, err := queries.CreateIndexerClient(ctx, params); err != nil {
		if err := handleWriteError(err, indexerClientErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r indexerClientRepository) ListByOwner(ctx context.Context, owner shared.UserID) ([]indexing.IndexerClient, error) {
	queries := sqlcgen.New(r.conn)

	entityModels, err := queries.ListIndexerClientsByOwner(ctx, owner.String())
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]indexing.IndexerClient, 0, len(entityModels))

	for i := range entityModels {
		item := entityModels[i]
		response = append(response, indexing.IndexerClient{
			ID:             shared.ParseID[indexing.IndexerID](item.ID),
			OwnerID:        shared.ParseID[shared.UserID](item.OwnerID),
			Type:           indexing.IndexerTypeProwlarr,
			Address:        *errutils.Must(url.Parse(item.Address)),
			Authentication: mappers.NewAuthentication(item.AuthCredentials),
		})
	}

	return response, nil
}

func (r indexerClientRepository) Update(ctx context.Context, id indexing.IndexerID, update func(indexerClient *indexing.IndexerClient) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	entityModel, err := queries.GetIndexerClient(ctx, id.String())
	if err != nil {
		return handleReadError(err)
	}

	indexerClient := &indexing.IndexerClient{
		ID:             shared.ParseID[indexing.IndexerID](entityModel.ID),
		OwnerID:        shared.ParseID[shared.UserID](entityModel.OwnerID),
		Type:           indexing.IndexerTypeProwlarr,
		Address:        *errutils.Must(url.Parse(entityModel.Address)),
		Authentication: mappers.NewAuthentication(entityModel.AuthCredentials),
	}

	if err := update(indexerClient); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateIndexerAddressParams{
		ID:      id.String(),
		Address: indexerClient.Address.String(),
	}

	if err = queries.UpdateIndexerAddress(ctx, updateParams); err != nil {
		if err := handleWriteError(err, indexerClientErrorHandler); err != nil {
			return err
		}

		return handleReadError(err)
	}

	updateAuthParams := sqlcgen.UpdateCredentialsParams{
		ID:          entityModel.AuthID,
		Credentials: mappers.NewAuthenticationModel(indexerClient.Authentication),
	}

	if err = queries.UpdateCredentials(ctx, updateAuthParams); err != nil {
		if err := handleWriteError(err, transferClientErrorHandler); err != nil {
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

	if err := queries.DeleteIndexerClient(ctx, id.String()); err != nil {
		return handleReadError(err)
	}

	return nil
}
