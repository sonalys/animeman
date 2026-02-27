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
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/ports"
	"github.com/sonalys/animeman/internal/utils/errgroup"
	"github.com/sonalys/animeman/internal/utils/otel"
	"google.golang.org/grpc/codes"
)

type (
	Repositories struct {
		ports.UserRepository
		ports.IndexerClientRepository
		ports.TransferClientRepository
		ports.CollectionRepository
		ports.TaskRepository
		ports.WatchlistRepository
		ports.FileRepository
	}

	Factories struct {
		ports.TransferClientControllerFactory
		ports.IndexingClientControllerFactory
	}

	usecases struct {
		repositories Repositories
		factories    Factories
	}

	Usecases interface {
		RegisterUser(ctx context.Context, username string, password string) (*users.User, error)
		Login(ctx context.Context, username string, password string) (*shared.UserID, error)

		CreateIndexer(ctx context.Context, args CreateIndexerArgs) (*indexing.Client, error)
		ListIndexers(ctx context.Context, userID shared.UserID) ([]indexing.Client, error)
		TestIndexingClientBuilder(ctx context.Context, b *indexing.ClientBuilder) error

		CreateTransferClient(ctx context.Context, args CreateTransferClientArgs) (*transfer.Client, error)
		ListTransferClients(ctx context.Context, userID shared.UserID) ([]transfer.Client, error)
		TestTransferClientBuilder(ctx context.Context, b *transfer.ClientBuilder) error

		GetOnboardingStatus(ctx context.Context, id shared.UserID) (*OnboardingStatus, error)
	}
)

func NewUsecases(
	r Repositories,
	f Factories,
) usecases {
	return usecases{
		repositories: r,
		factories:    f,
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
	Type   indexing.ClientType
	URL    url.URL
	Auth   authentication.Authentication
	UserID shared.UserID
}

func (u usecases) CreateIndexer(ctx context.Context, args CreateIndexerArgs) (*indexing.Client, error) {
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

type CreateTransferClientArgs struct {
	Type   transfer.ClientType
	URL    url.URL
	Auth   authentication.Authentication
	UserID shared.UserID
}

func (u usecases) CreateTransferClient(ctx context.Context, args CreateTransferClientArgs) (*transfer.Client, error) {
	ctx, span := otel.Tracer.Start(ctx, "CreateTransferClient")
	defer span.End()

	client, err := transfer.NewClient(
		args.UserID,
		args.Type,
		args.URL,
		args.Auth,
	)
	if err != nil {
		logError(ctx, err, "Failed creating transfer client")
		return nil, apperr.New(err, codes.InvalidArgument)
	}

	if err := u.repositories.TransferClientRepository.Create(ctx, client); err != nil {
		logError(ctx, err, "Failed creating transfer client")
		return nil, apperr.New(err, codes.InvalidArgument)
	}

	log.Info().
		Ctx(ctx).
		Msg("Created transfer client")

	return client, nil
}

func (u usecases) ListIndexers(ctx context.Context, userID shared.UserID) ([]indexing.Client, error) {
	ctx, span := otel.Tracer.Start(ctx, "ListIndexers")
	defer span.End()

	response, err := u.repositories.IndexerClientRepository.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (u usecases) ListTransferClients(ctx context.Context, userID shared.UserID) ([]transfer.Client, error) {
	ctx, span := otel.Tracer.Start(ctx, "ListIndexers")
	defer span.End()

	response, err := u.repositories.TransferClientRepository.ListByOwner(ctx, userID)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func (u usecases) TestTransferClientBuilder(ctx context.Context, b *transfer.ClientBuilder) error {
	client, err := b.Build()
	if err != nil {
		logError(ctx, err, "creating transfer client")
		return err
	}

	if _, err = u.factories.TransferClientControllerFactory.New(ctx, client); err != nil {
		logError(ctx, err, "Error testing transfer client configuration")
		return err
	}

	return nil
}

func (u usecases) TestIndexingClientBuilder(ctx context.Context, b *indexing.ClientBuilder) error {
	client, err := b.Build()
	if err != nil {
		logError(ctx, err, "creating indexing client")
		return err
	}

	if _, err = u.factories.IndexingClientControllerFactory.New(ctx, client); err != nil {
		logError(ctx, err, "Error testing indexing client configuration")
		return err
	}

	return nil
}

type SetupStep string

const (
	SetupStepTransferClient     SetupStep = "transferClient"
	SetupStepIndexingClient     SetupStep = "indexingClient"
	SetupStepWatchlistSetupStep SetupStep = "watchlist"
)

type OnboardingStatus struct {
	IsSetupCompleted bool
	OptionalSteps    []SetupStep
	MissingSteps     []SetupStep
	CompletedSteps   []SetupStep
}

func (u usecases) GetOnboardingStatus(ctx context.Context, id shared.UserID) (*OnboardingStatus, error) {
	status := &OnboardingStatus{}

	isCompleted, err := u.repositories.UserRepository.IsSetupCompleted(ctx, id)
	if err != nil {
		logError(ctx, err, "Failed to retrieve user setup completedness")
		return nil, err
	}

	if isCompleted {
		status.IsSetupCompleted = isCompleted
		return status, nil
	}

	errgrp, grpctx := errgroup.WithContext(ctx)

	errgrp.Go(func() error {
		entries, err := u.repositories.TransferClientRepository.List(grpctx)

		if err == nil {
			if len(entries) > 0 {
				status.CompletedSteps = append(status.CompletedSteps, SetupStepTransferClient)
			} else {
				status.MissingSteps = append(status.MissingSteps, SetupStepTransferClient)
			}
		}

		return err
	})

	errgrp.Go(func() error {
		entries, err := u.repositories.IndexerClientRepository.List(grpctx)

		if err == nil {
			if len(entries) > 0 {
				status.CompletedSteps = append(status.CompletedSteps, SetupStepIndexingClient)
			} else {
				status.MissingSteps = append(status.MissingSteps, SetupStepIndexingClient)
			}
		}

		return err
	})

	errgrp.Go(func() error {
		entries, err := u.repositories.WatchlistRepository.List(grpctx)

		if err == nil {
			if len(entries) > 0 {
				status.CompletedSteps = append(status.CompletedSteps, SetupStepWatchlistSetupStep)
			} else {
				status.MissingSteps = append(status.MissingSteps, SetupStepWatchlistSetupStep)
			}
		}

		return err
	})

	if err := errgrp.Wait(); err != nil {
		logError(ctx, err, "Failed to retrieve onboarding status")
		return nil, err
	}

	return status, nil
}
