package users

import (
	"fmt"
	"net/url"
	"slices"
	"time"

	"github.com/sonalys/animeman/internal/app/apperr"
	"github.com/sonalys/animeman/internal/domain/collections"
	"github.com/sonalys/animeman/internal/domain/shared"
	"github.com/sonalys/animeman/internal/domain/transfer"
	"github.com/sonalys/animeman/internal/domain/watchlists"
	"golang.org/x/crypto/bcrypt"
	"google.golang.org/grpc/codes"
)

type (
	User struct {
		ID           shared.UserID
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
		ID:           shared.NewID[shared.UserID](),
		Username:     username,
		PasswordHash: hashedPassword,
	}, nil
}

func (u User) NewProwlarrConfiguration(
	host string,
	apiKey string,
) *collections.ProwlarrConfiguration {
	return &collections.ProwlarrConfiguration{
		ID:      shared.NewID[collections.ProwlarrConfigID](),
		OwnerID: u.ID,
		Host:    host,
		APIKey:  apiKey,
	}
}

func (u User) NewTransferClient(
	clientType transfer.ClientType,
	address url.URL,
	auth transfer.Authentication,
) *transfer.Client {
	return &transfer.Client{
		ID:             shared.NewID[transfer.ClientID](),
		OwnerID:        u.ID,
		Type:           clientType,
		Address:        address,
		Authentication: auth,
	}
}

func (u User) NewExternalWatchList(
	source watchlists.WatchlistSource,
	externalID string,
	syncFrequency time.Duration,
) *watchlists.Watchlist {
	return &watchlists.Watchlist{
		ID:            shared.NewID[watchlists.WatchlistID](),
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
) *collections.Collection {
	return &collections.Collection{
		ID:        shared.NewID[collections.CollectionID](),
		Owner:     u.ID,
		Name:      name,
		BasePath:  basePath,
		Tags:      slices.Compact(tags),
		Monitored: monitored,
		CreatedAt: time.Now(),
	}
}
