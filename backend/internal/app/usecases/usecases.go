package usecases

import (
	"context"

	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/users"
	"github.com/sonalys/animeman/internal/ports"
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
	newUser, err := users.NewUser(username, []byte(password))
	if err != nil {
		return nil, err
	}

	if err := u.repositories.UserRepository.Create(ctx, newUser); err != nil {
		return nil, err
	}

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
