package repositories

import (
	"context"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/ports"
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
	panic("unimplemented")
}

func (r *mediaRepository) Delete(ctx context.Context, id collections.MediaID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteMedia(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r *mediaRepository) ListByCollection(ctx context.Context, id collections.CollectionID, opts ports.ListOptions) ([]collections.Media, error) {
	panic("unimplemented")
}

func (r *mediaRepository) Update(ctx context.Context, id collections.MediaID, updateHandler ports.UpdateHandler[collections.Media]) error {
	panic("unimplemented")
}
