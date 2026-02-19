package usecases

import (
	"context"
	"errors"
	"net/url"

	"github.com/rs/zerolog/log"
	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/indexing"
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
		Login(ctx context.Context, username string, password string) (*shared.UserID, error)
		CreateIndexer(ctx context.Context, args CreateIndexerArgs) (*indexing.IndexerClient, error)
		ListIndexers(ctx context.Context, userID shared.UserID) ([]indexing.IndexerClient, error)
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
		logError(ctx, err, "Failed creating new user")
		return nil, err
	}

	span.AddEvent("User created")

	if err := u.repositories.UserRepository.Create(ctx, newUser); err != nil {
		logError(ctx, err, "Failed registering new user")

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

func (u usecases) Login(ctx context.Context, username string, password string) (*shared.UserID, error) {
	ctx, span := otel.Tracer.Start(ctx, "Login")
	defer span.End()

	user, err := u.repositories.UserRepository.GetByUsername(ctx, username)
	if err != nil {
		logError(ctx, err, "Failed retrieving user")
		return nil, err
	}

	if err := user.Login([]byte(password)); err != nil {
		logError(ctx, err, "Failed login user")
		return nil, apperr.New(err, codes.InvalidArgument)
	}

	log.Info().
		Ctx(ctx).
		Msg("User logged in")

	return &user.ID, nil
}

type CreateIndexerArgs struct {
	Type   indexing.IndexerType
	URL    url.URL
	Auth   authentication.Authentication
	UserID shared.UserID
}

func (u usecases) CreateIndexer(ctx context.Context, args CreateIndexerArgs) (*indexing.IndexerClient, error) {
	ctx, span := otel.Tracer.Start(ctx, "CreateIndexer")
	defer span.End()

	client := indexing.NewClient(
		args.UserID,
		args.Type,
		args.URL,
		args.Auth,
	)

	if err := u.repositories.IndexerClientRepository.Create(ctx, client); err != nil {
		logError(ctx, err, "Failed creating indexer client")
		return nil, apperr.New(err, codes.InvalidArgument)
	}

	log.Info().
		Ctx(ctx).
		Msg("Created indexer client")

	return client, nil
}

func (u usecases) ListIndexers(ctx context.Context, userID shared.UserID) ([]indexing.IndexerClient, error) {
	ctx, span := otel.Tracer.Start(ctx, "ListIndexers")
	defer span.End()

	response, err := u.repositories.IndexerClientRepository.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}
