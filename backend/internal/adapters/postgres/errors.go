package postgres

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonalys/animeman/internal/app/apperr"
	"google.golang.org/grpc/codes"
)

func handleWriteError(err error, handler func(err *pgconn.PgError) error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		if err := handler(pgErr); err != nil {
			return err
		}
	}

	return nil
}

func handleReadError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return apperr.New(err, codes.NotFound)
	default:
		return err
	}
}
