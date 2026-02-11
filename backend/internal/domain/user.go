package domain

import (
	"fmt"

	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           string
	Username     string
	PasswordHash []byte
}

func NewUser(
	id string,
	email string,
	password []byte,
) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing user password: %w", err)
	}

	return &User{
		ID:           id,
		Username:     email,
		PasswordHash: hashedPassword,
	}, nil
}
