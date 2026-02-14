package repositories

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/ports"
	"google.golang.org/grpc/codes"
)

type collectionRepository struct {
	conn *pgxpool.Pool
}

func NewCollectionRepository(conn *pgxpool.Pool) ports.CollectionRepository {
	return &collectionRepository{
		conn: conn,
	}
}

func (r *collectionRepository) Create(ctx context.Context, collection *collections.Collection) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateCollectionParams{
		ID:        collection.ID,
		OwnerID:   collection.Owner,
		Name:      collection.Name,
		BasePath:  collection.BasePath,
		Tags:      collection.Tags,
		Monitored: collection.Monitored,
		CreatedAt: pgtype.Timestamptz{Time: collection.CreatedAt},
	}

	if _, err := queries.CreateCollection(ctx, params); err != nil {
		if err := handleWriteError(err, collectionErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r *collectionRepository) Delete(ctx context.Context, id collections.CollectionID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteCollection(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *collectionRepository) ListByOwner(ctx context.Context, owner shared.UserID) ([]collections.Collection, error) {
	queries := sqlcgen.New(r.conn)

	entityModels, err := queries.ListCollectionsByOwner(ctx, owner)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Collection, 0, len(entityModels))

	for i := range entityModels {
		item := entityModels[i]
		response = append(response, collections.Collection{
			ID:        item.ID,
			Owner:     item.OwnerID,
			Name:      item.Name,
			BasePath:  item.BasePath,
			Tags:      item.Tags,
			Monitored: item.Monitored,
			CreatedAt: item.CreatedAt.Time,
		})
	}

	return response, nil
}

func (r *collectionRepository) Update(ctx context.Context, id collections.CollectionID, update func(collection *collections.Collection) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	entityModel, err := queries.GetCollection(ctx, id)
	if err != nil {
		return handleReadError(err)
	}

	collection := &collections.Collection{
		ID:        entityModel.ID,
		Owner:     entityModel.OwnerID,
		Name:      entityModel.Name,
		BasePath:  entityModel.BasePath,
		Tags:      entityModel.Tags,
		Monitored: entityModel.Monitored,
		CreatedAt: entityModel.CreatedAt.Time,
	}

	if err := update(collection); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateCollectionParams{
		ID:        id,
		Name:      collection.Name,
		BasePath:  collection.BasePath,
		Tags:      collection.Tags,
		Monitored: collection.Monitored,
	}

	if _, err = queries.UpdateCollection(ctx, updateParams); err != nil {
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

func collectionErrorHandler(err *pgconn.PgError) error {
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
