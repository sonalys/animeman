package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain"
)

func NewUser(from *sqlcgen.User) *domain.User {
	return &domain.User{
		ID:           domain.ParseID[domain.UserID](from.ID),
		Username:     from.Username,
		PasswordHash: []byte(from.PasswordHash),
	}
}
