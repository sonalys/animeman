package ports

import (
	"context"

	"github.com/sonalys/animeman/internal/domain"
)

type (
	UserRepository interface {
		Create(ctx context.Context, user *domain.User) error
		Get(ctx context.Context, id string) (*domain.User, error)
		Update(ctx context.Context, id string, update func(user *domain.User) error) error
		Delete(ctx context.Context, id string) error
	}
)
