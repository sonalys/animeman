package repositories

import (
	"context"
	"net/url"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/indexing"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/errutils"
)

type indexerClientRepository struct {
	conn *pgxpool.Pool
}

func NewIndexerClientRepository(conn *pgxpool.Pool) ports.IndexerClientRepository {
	return &indexerClientRepository{
		conn: conn,
	}
}

func (r indexerClientRepository) Create(ctx context.Context, client *indexing.Client) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		auth, err := queries.CreateAuthentication(ctx, sqlcgen.CreateAuthenticationParams{
			ID:          shared.NewID[shared.ID](),
			Type:        mappers.NewAuthenticationTypeModel(client.Authentication.Type),
			Credentials: mappers.NewAuthenticationModel(client.Authentication),
		})

		if err != nil {
			return handleWriteError(err)
		}

		params := sqlcgen.CreateIndexerClientParams{
			ID:      client.ID,
			OwnerID: client.OwnerID,
			Address: client.Address.String(),
			Type:    sqlcgen.IndexerClientType(client.Type.String()),
			AuthID:  auth.ID,
		}

		if _, err := queries.CreateIndexerClient(ctx, params); err != nil {
			return handleWriteError(err)
		}

		return nil
	})
}

func (r indexerClientRepository) List(ctx context.Context) ([]indexing.Client, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListIndexerClients(ctx)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]indexing.Client, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, indexing.Client{
			ID:             model.ID,
			OwnerID:        model.OwnerID,
			Type:           indexing.IndexerTypeProwlarr,
			Address:        *errutils.Must(url.Parse(model.Address)),
			Authentication: mappers.NewAuthentication(model.AuthCredentials),
		})
	}

	return response, nil
}

func (r indexerClientRepository) ListByOwner(ctx context.Context, owner shared.UserID) ([]indexing.Client, error) {
	queries := sqlcgen.New(r.conn)

	models, err := queries.ListIndexerClientsByOwner(ctx, owner)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]indexing.Client, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, indexing.Client{
			ID:             model.ID,
			OwnerID:        model.OwnerID,
			Type:           indexing.IndexerTypeProwlarr,
			Address:        *errutils.Must(url.Parse(model.Address)),
			Authentication: mappers.NewAuthentication(model.AuthCredentials),
		})
	}

	return response, nil
}

func (r indexerClientRepository) Update(ctx context.Context, id indexing.IndexerID, update func(indexerClient *indexing.Client) error) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetIndexerClient(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		indexerClient := &indexing.Client{
			ID:             model.ID,
			OwnerID:        model.OwnerID,
			Type:           indexing.IndexerTypeProwlarr,
			Address:        *errutils.Must(url.Parse(model.Address)),
			Authentication: mappers.NewAuthentication(model.AuthCredentials),
		}

		if err := update(indexerClient); err != nil {
			return err
		}

		updateParams := sqlcgen.UpdateIndexerAddressParams{
			ID:      id,
			Address: indexerClient.Address.String(),
		}

		if err = queries.UpdateIndexerAddress(ctx, updateParams); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		updateAuthParams := sqlcgen.UpdateCredentialsParams{
			ID:          model.AuthID,
			Credentials: mappers.NewAuthenticationModel(indexerClient.Authentication),
		}

		if err = queries.UpdateCredentials(ctx, updateAuthParams); err != nil {
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}

func (r indexerClientRepository) Delete(ctx context.Context, id indexing.IndexerID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteIndexerClient(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}
