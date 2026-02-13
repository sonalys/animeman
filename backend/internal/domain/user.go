package domain

import (
	"fmt"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/app/apperr"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type (
	UserID struct{ uuid.UUID }

	User struct {
		ID           UserID
		Username     string
		PasswordHash []byte
	}
)

func NewUser(
	username string,
	password []byte,
) (*User, error) {
	if pwdLength := len(password); pwdLength < 8 || pwdLength > 72 {
		return nil, apperr.New(ErrInvalidPasswordLength, codes.InvalidArgument)
	}

	hashedPassword, err := bcrypt.GenerateFromPassword(password, bcrypt.DefaultCost)
	if err != nil {
		return nil, fmt.Errorf("hashing user password: %w", err)
	}

	return &User{
		ID:           NewID[UserID](),
		Username:     username,
		PasswordHash: hashedPassword,
	}, nil
}

func (u User) NewProwlarrConfiguration(
	host string,
	apiKey string,
) *ProwlarrConfiguration {
	return &ProwlarrConfiguration{
		ID:      NewID[ProwlarrConfigID](),
		OwnerID: u.ID,
		Host:    host,
		APIKey:  apiKey,
	}
}

func (u User) NewTorrentClientConfiguration(
	source TorrentSource,
	host string,
	username string,
	password []byte,
) *TorrentClient {
	return &TorrentClient{
		ID:       NewID[TorrentClientID](),
		OwnerID:  u.ID,
		Source:   source,
		Host:     host,
		Username: username,
		Password: password,
	}
}
