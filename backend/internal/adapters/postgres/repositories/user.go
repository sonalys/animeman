package repositories

import (
	"context"

	"github.com/jackc/pgerrcode"
	"github.com/jackc/pgx/v5/pgconn"
	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/sonalys/animeman/internal/adapters/postgres/mappers"
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/ports"
	"google.golang.org/grpc/codes"
)

type userRepository struct {
	conn *pgxpool.Pool
}

func NewUserRepository(conn *pgxpool.Pool) ports.UserRepository {
	return &userRepository{
		conn: conn,
	}
}

func (r userRepository) Create(ctx context.Context, user *users.User) error {
	queries := sqlcgen.New(r.conn)

	params := sqlcgen.CreateUserParams{
		ID:           user.ID,
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

func (r userRepository) Delete(ctx context.Context, id shared.UserID) error {
	queries := sqlcgen.New(r.conn)

	if err := queries.DeleteUser(ctx, id); err != nil {
		return handleReadError(err)
	}

	return nil
}

func (r userRepository) Get(ctx context.Context, id shared.UserID) (*users.User, error) {
	queries := sqlcgen.New(r.conn)

	userModel, err := queries.GetUserById(ctx, id)
	if err != nil {
		return nil, handleReadError(err)
	}

	user := mappers.NewUser(&userModel)

	return user, nil
}

func (r userRepository) Update(ctx context.Context, id shared.UserID, update func(user *users.User) error) error {
	return transaction(ctx, r.conn, func(queries *sqlcgen.Queries) error {
		userModel, err := queries.GetUserById(ctx, id)
		if err != nil {
			return handleReadError(err)
		}

		user := mappers.NewUser(&userModel)

		if err := update(user); err != nil {
			return err
		}

		updateParams := sqlcgen.UpdateUserPasswordParams{
			ID:           user.ID,
			PasswordHash: string(user.PasswordHash),
		}

		if _, err = queries.UpdateUserPassword(ctx, updateParams); err != nil {
			if err := handleWriteError(err, userErrorHandler); err != nil {
				return err
			}

			return handleReadError(err)
		}

		return nil
	})
}

func userErrorHandler(err *pgconn.PgError) error {
	switch err.Code {
	case pgerrcode.UniqueViolation:
		switch err.ConstraintName {
		case "users_pkey":
			return apperr.New(users.ErrUniqueUsername, codes.AlreadyExists)
		default:
			return apperr.New(err, codes.FailedPrecondition)
		}
	default:
		return err
	}
}
