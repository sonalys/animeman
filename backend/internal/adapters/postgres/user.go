package postgres

import (
	"context"
	"database/sql"
	"errors"

	"github.com/gofrs/uuid/v5"
	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain"
	"google.golang.org/grpc/codes"
)

type userRepository struct {
	conn *pgxpool.Pool
}

func (r userRepository) Create(ctx context.Context, user *domain.User) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateUserParams{
		ID:           user.ID.String(),
		Username:     user.Username,
		PasswordHash: string(user.PasswordHash),
	}

	_, err := queries.CreateUser(ctx, params)
	if err != nil {
		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "unique_username":
				return apperr.New(err, codes.AlreadyExists, "username already exists")
			}
		}
		return apperr.New(err, codes.Internal, "could not create user")
	}

	return nil
}

func (r userRepository) Delete(ctx context.Context, id domain.UserID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteUser(ctx, id.String()); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "not found")
		default:
			return apperr.New(err, codes.Internal, "internal error")
		}
	}

	return nil
}

func (r userRepository) Get(ctx context.Context, id domain.UserID) (*domain.User, error) {
	queries := sqlcgen.New(r.conn)

	userModel, err := queries.GetUserById(ctx, id.String())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, apperr.New(err, codes.NotFound, "not found")
		default:
			return nil, apperr.New(err, codes.Internal, "internal error")
		}
	}

	user := &domain.User{
		ID:           domain.UserID{uuid.FromStringOrNil(userModel.ID)},
		Username:     userModel.Username,
		PasswordHash: []byte(userModel.PasswordHash),
	}

	return user, nil
}

func (r userRepository) Update(ctx context.Context, id domain.UserID, update func(user *domain.User) error) error {
	tx, err := r.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return apperr.New(err, codes.Internal, "could not start transaction")
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	userModel, err := queries.GetUserById(ctx, id.String())
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "not found")
		default:
			return apperr.New(err, codes.Internal, "internal error")
		}
	}

	user := &domain.User{
		ID:           domain.UserID{uuid.FromStringOrNil(userModel.ID)},
		Username:     userModel.Username,
		PasswordHash: []byte(userModel.PasswordHash),
	}

	if err := update(user); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateUserPasswordParams{
		ID:           user.ID.String(),
		PasswordHash: string(user.PasswordHash),
	}

	if _, err = queries.UpdateUserPassword(ctx, updateParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.New(err, codes.NotFound, "not found")
		}

		if pgErr, ok := errors.AsType[*pgconn.PgError](err); ok && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "unique_username":
				return apperr.New(err, codes.AlreadyExists, "username already exists")
			}
		}
		return apperr.New(err, codes.Internal, "could not create user")
	}

	if err := tx.Commit(ctx); err != nil {
		return apperr.New(err, codes.Internal, "could not commit transaction")
	}

	return nil
}
