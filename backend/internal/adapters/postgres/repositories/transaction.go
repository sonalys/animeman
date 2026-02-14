package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
)

func transaction(ctx context.Context, conn *pgxpool.Pool, handler func(queries *sqlcgen.Queries) error) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	if err := handler(queries); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
