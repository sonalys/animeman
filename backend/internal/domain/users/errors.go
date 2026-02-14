package users

import "github.com/sonalys/animeman/internal/domain/shared"

const (
	ErrUniqueUsername        shared.StringError = "username must be unique"
	ErrInvalidPasswordLength shared.StringError = "password must be between 8 and 72 digits"
)
