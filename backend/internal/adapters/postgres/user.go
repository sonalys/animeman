package postgres

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain"
	"google.golang.org/grpc/codes"
)

type userRepository struct {
	conn *pgxpool.Pool
}

func userErrorHandler(err *pgconn.PgError) error {
	switch err.Code {
	case pgerrcode.UniqueViolation:
		switch err.ConstraintName {
		case "users_pkey":
			return apperr.New(domain.ErrUniqueUsername, codes.AlreadyExists)
		default:
			return apperr.New(err, codes.FailedPrecondition)
		}
	default:
		return err
	}
}

func (r userRepository) Create(ctx context.Context, user *domain.User) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateUserParams{
		ID:           user.ID.String(),
		Username:     user.Username,
		PasswordHash: string(user.PasswordHash),
	}

	if _, err := queries.CreateUser(ctx, params); err != nil {
		if err := handleWriteError(err, userErrorHandler); err != nil {
			return err
		}

		return err
	}

	return nil
}

func (r userRepository) Delete(ctx context.Context, id domain.UserID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteUser(ctx, id.String()); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r userRepository) Get(ctx context.Context, id domain.UserID) (*domain.User, error) {
	queries := sqlcgen.New(r.conn)

	userModel, err := queries.GetUserById(ctx, id.String())
	if err != nil {
		return nil, handleReadError(err)
	}

	user := mappers.NewUser(&userModel)

	return user, nil
}

func (r userRepository) Update(ctx context.Context, id domain.UserID, update func(user *domain.User) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return err
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	userModel, err := queries.GetUserById(ctx, id.String())
	if err != nil {
		return handleReadError(err)
	}

	user := mappers.NewUser(&userModel)

	if err := update(user); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateUserPasswordParams{
		ID:           user.ID.String(),
		PasswordHash: string(user.PasswordHash),
	}

	if _, err = queries.UpdateUserPassword(ctx, updateParams); err != nil {
		if err := handleWriteError(err, userErrorHandler); err != nil {
			return err
		}

		return handleReadError(err)
	}

	if err := tx.Commit(ctx); err != nil {
		return err
	}

	return nil
}
