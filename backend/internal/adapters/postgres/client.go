package postgres

import (
	"context"
	"fmt"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/postgres/repositories"
	"github.com/sonalys/animeman/internal/ports"
)

type Client struct {
	conn *pgxpool.Pool
}

func New(ctx context.Context, connStr string) (*Client, error) {
	cfg, err := pgxpool.ParseConfig(connStr)
	if err != nil {
		return nil, fmt.Errorf("failed to parse connStr: %w", err)
	}

	cfg.ConnConfig.Tracer = tracer{}

	dbpool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create connection pool: %w", err)
	}

	if err := waitConnection(ctx, dbpool); err != nil {
		return nil, fmt.Errorf("waiting for postgres connection: %w", err)
	}

	return &Client{
		conn: dbpool,
	}, nil
}

func (c Client) UserRepository() ports.UserRepository {
	return repositories.NewUserRepository(c.conn)
}

func (c Client) IndexerClientRepository() ports.IndexerClientRepository {
	return repositories.NewIndexerClientRepository(c.conn)
}

func (c Client) TransferClientRepository() ports.TransferClientRepository {
	return repositories.NewTransferClientRepository(c.conn)
}

func (c Client) CollectionRepository() ports.CollectionRepository {
	return repositories.NewCollectionRepository(c.conn)
}

func waitConnection(ctx context.Context, conn *pgxpool.Pool) error {
	for {
		if err := conn.Ping(ctx); err == nil {
			return nil
		}

		log.Trace().Msg("Waiting for postgres connection")

		select {
		case <-ctx.Done():
			return ctx.Err()
		case <-time.After(500 * time.Millisecond):
		}
	}
}
