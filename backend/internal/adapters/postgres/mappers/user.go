package mappers

import (
	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/users"
)

func NewUser(from *sqlcgen.User) *users.User {
	return &users.User{
		ID:           from.ID,
		Username:     from.Username,
		PasswordHash: []byte(from.PasswordHash),
	}
}
