package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain"
)

type (
	UserRepository interface {
		Create(ctx context.Context, user *domain.User) error
		Get(ctx context.Context, id domain.UserID) (*domain.User, error)
		Update(ctx context.Context, id domain.UserID, update func(user *domain.User) error) error
		Delete(ctx context.Context, id domain.UserID) error
	}

	ProwlarrRepository interface {
		CreateConfig(ctx context.Context, config *domain.ProwlarrConfiguration) error
		GetConfigByOwner(ctx context.Context, owner domain.UserID) (*domain.ProwlarrConfiguration, error)
		UpdateConfig(ctx context.Context, id domain.ProwlarrConfigID, update func(config *domain.ProwlarrConfiguration) error) error
		DeleteConfig(ctx context.Context, id domain.ProwlarrConfigID) error
	}
)
