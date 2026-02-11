package domain

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID           uuid.UUID
	Username     string
	PasswordHash []byte
}

func NewUser(
	email string,
	password []byte,
) (*User, error) {
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing user password: %w", err)
	}

	return &User{
		ID:           uuid.Must(uuid.NewV7()),
		Username:     email,
		PasswordHash: hashedPassword,
	}, nil
}

func (u User) CreateProwlarrConfiguration(
	host string,
	apiKey string,
) *ProwlarrConfiguration {
	return &ProwlarrConfiguration{
		ID:      uuid.Must(uuid.NewV7()),
		OwnerID: u.ID,
		Host:    host,
		APIKey:  apiKey,
	}
}
