package repositories

import (
	"context"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgxpool"
)

func transaction(ctx context.Context, conn *pgxpool.Pool, handler func(tx pgx.Tx) error) error {
	tx, err := conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	if err := handler(tx); err != nil {
		return err
	}

	return tx.Commit(ctx)
}
