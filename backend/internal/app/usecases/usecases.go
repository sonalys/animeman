package usecases

import (
	"context"
	"errors"

	"github.com/rs/zerolog"
	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/otel"
	"google.golang.org/grpc/codes"
)

type (
	Repositories struct {
		ports.UserRepository
		ports.IndexerClientRepository
		ports.TransferClientRepository
		ports.CollectionRepository
		ports.WatchlistRepository
	}

	usecases struct {
		repositories Repositories
	}

	Usecases interface {
		RegisterUser(ctx context.Context, username string, password string) (*users.User, error)
		Login(ctx context.Context, username string, password []byte) (*shared.UserID, error)
	}
)

func NewUsecases(r Repositories) usecases {
	return usecases{
		repositories: r,
	}
}

func (u usecases) RegisterUser(ctx context.Context, username string, password string) (*users.User, error) {
	ctx, span := otel.Tracer.Start(ctx, "RegisterUser")
	defer span.End()

	newUser, err := users.NewUser(username, []byte(password))
	if err != nil {
		logError(ctx, err, "Could not initialize new user")
		return nil, err
	}

	span.AddEvent("User created")

	if err := u.repositories.UserRepository.Create(ctx, newUser); err != nil {
		logError(ctx, err, "Could not register new user")

		if errors.Is(err, users.ErrUniqueUsername) {
			return nil, apperr.NewPublicError(err, "username '%s' already exists", newUser.Username)
		}
		return nil, err
	}

	log.Info().
		Ctx(ctx).
		Msg("User registered")

	return newUser, nil
}

func (u usecases) Login(ctx context.Context, username string, password []byte) (*shared.UserID, error) {
	user, err := u.repositories.UserRepository.GetByUsername(ctx, username)
	if err != nil {
		return nil, err
	}

	if err := user.Login(password); err != nil {
		return nil, apperr.New(err, codes.InvalidArgument)
	}

	return &user.ID, nil
}

func logError(ctx context.Context, err error, mask string, args ...any) {
	level := zerolog.ErrorLevel

	appErr, ok := errors.AsType[apperr.Error](err)
	if ok {
		if appErr.Code() != codes.Internal {
			level = zerolog.InfoLevel
		}

		log.
			WithLevel(level).
			Ctx(ctx).
			Stringer("code", appErr.Code()).
			Str("message", appErr.Message).
			Err(appErr.Cause).
			Msgf(mask, args...)
		return
	}

	log.
		WithLevel(level).
		Ctx(ctx).
		Err(err).
		Msgf(mask, args...)
}
