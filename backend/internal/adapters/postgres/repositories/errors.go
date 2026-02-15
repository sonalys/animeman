package repositories

import (
	"database/sql"
	"errors"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/sonalys/animeman/internal/app/apperr"
	"google.golang.org/grpc/codes"
)

func handleWriteError(err error, handlers ...func(err *pgconn.PgError) error) error {
	if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok {
		for _, handler := range handlers {
			if err := handler(pgErr); err != nil {
				return err
			}
		}

		switch pgErr.Code {
		case pgerrcode.UniqueViolation:
			switch pgErr.ConstraintName {
			default:
				return apperr.New(err, codes.FailedPrecondition)
			}
		}
	}

	return err
}

func handleReadError(err error) error {
	switch {
	case errors.Is(err, sql.ErrNoRows):
		return apperr.New(err, codes.NotFound)
	default:
		return err
	}
}
