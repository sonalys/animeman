package domain

import (
	"fmt"
	"slices"
	"time"

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

func (u User) NewExternalWatchList(
	source WatchlistSource,
	externalID string,
	syncFrequency time.Duration,
) *Watchlist {
	return &Watchlist{
		ID:            NewID[WatchlistID](),
		Owner:         u.ID,
		Source:        source,
		ExternalID:    externalID,
		SyncFrequency: syncFrequency,
		CreatedAt:     time.Now(),
	}
}

func (u User) NewCollection(
	name string,
	basePath string,
	tags []string,
	monitored bool,
) *Collection {
	return &Collection{
		ID:        NewID[CollectionID](),
		Owner:     u.ID,
		Name:      name,
		BasePath:  basePath,
		Tags:      slices.Compact(tags),
		Monitored: monitored,
		CreatedAt: time.Now(),
	}
}
