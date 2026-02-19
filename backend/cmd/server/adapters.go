package main

import (
	"context"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/adapters/postgres"
)

type adapters struct {
	postgresClient *postgres.Client
}

func initializeAdapters(
	ctx context.Context,
	postgresConn string,
) adapters {
	postgresClient, err := postgres.New(ctx, postgresConn)
	if err != nil {
		log.Fatal().
			Err(err).
			Msg("Could not connect to PostgreSQL")
	}

	return adapters{
		postgresClient: postgresClient,
	}
}
