package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/ports"
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
		return handleWriteError(err)
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

	models, err := queries.ListCollectionsByOwner(ctx, owner)
	if err != nil {
		return nil, handleReadError(err)
	}

	response := make([]collections.Collection, 0, len(models))

	for i := range models {
		model := &models[i]

		response = append(response, collections.Collection{
			ID:        model.ID,
			Owner:     model.OwnerID,
			Name:      model.Name,
			BasePath:  model.BasePath,
			Tags:      model.Tags,
			Monitored: model.Monitored,
			CreatedAt: model.CreatedAt.Time,
		})
	}

	return response, nil
}

func (r *collectionRepository) Update(ctx context.Context, id collections.CollectionID, update func(collection *collections.Collection) error) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		model, err := queries.GetCollection(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		collection := &collections.Collection{
			ID:        model.ID,
			Owner:     model.OwnerID,
			Name:      model.Name,
			BasePath:  model.BasePath,
			Tags:      model.Tags,
			Monitored: model.Monitored,
			CreatedAt: model.CreatedAt.Time,
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
			if err := handleWriteError(err); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}
