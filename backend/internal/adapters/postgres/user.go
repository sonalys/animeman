package postgres

import (
	"context"
	"database/sql"
	"errors"

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

func (u userRepository) Create(ctx context.Context, user *domain.User) error {
	queries := sqlcgen.New(u.conn)

	params := sqlcgen.CreateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: string(user.PasswordHash),
	}

	err := queries.CreateUser(ctx, params)
	if err != nil {
		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
			switch pgErr.ConstraintName {
			case "unique_username":
				return apperr.New(err, codes.AlreadyExists, "username already exists")
			}
		}
		return apperr.New(err, codes.Internal, "could not create user")
	}

	return nil
}

func (u userRepository) Delete(ctx context.Context, id string) error {
	queries := sqlcgen.New(u.conn)

	if err := queries.DeleteUser(ctx, id); err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "user not found")
		default:
			return apperr.New(err, codes.Internal, "could not get user")
		}
	}

	return nil
}

func (u userRepository) Get(ctx context.Context, id string) (*domain.User, error) {
	queries := sqlcgen.New(u.conn)

	userModel, err := queries.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return nil, apperr.New(err, codes.NotFound, "user not found")
		default:
			return nil, apperr.New(err, codes.Internal, "could not get user")
		}
	}

	user := &domain.User{
		ID:           userModel.ID,
		Username:     userModel.Username,
		PasswordHash: []byte(userModel.PasswordHash),
	}

	return user, nil
}

func (u userRepository) Update(ctx context.Context, id string, update func(user *domain.User) error) error {
	tx, err := u.conn.BeginTx(ctx, pgx.TxOptions{})
	if err != nil {
		return apperr.New(err, codes.Internal, "could not start transaction")
	}
	defer tx.Rollback(ctx)

	queries := sqlcgen.New(tx)

	userModel, err := queries.GetUser(ctx, id)
	if err != nil {
		switch {
		case errors.Is(err, sql.ErrNoRows):
			return apperr.New(err, codes.NotFound, "user not found")
		default:
			return apperr.New(err, codes.Internal, "could not get user")
		}
	}

	user := &domain.User{
		ID:           userModel.ID,
		Username:     userModel.Username,
		PasswordHash: []byte(userModel.PasswordHash),
	}

	if err := update(user); err != nil {
		return err
	}

	updateParams := sqlcgen.UpdateUserParams{
		ID:           user.ID,
		Username:     user.Username,
		PasswordHash: string(user.PasswordHash),
	}

	if err = queries.UpdateUser(ctx, updateParams); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return apperr.New(err, codes.NotFound, "user not found")
		}

		var pgErr *pgconn.PgError
		if errors.As(err, &pgErr) && pgErr.Code == "23505" {
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
