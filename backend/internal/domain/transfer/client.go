package transfer

import (
	"net/url"

	"github.com/gofrs/uuid/v5"
	"github.com/sonalys/animeman/internal/domain/authentication"
	"github.com/sonalys/animeman/internal/domain/hashing"
	"github.com/sonalys/animeman/internal/domain/releases"
	"github.com/sonalys/animeman/internal/domain/shared"
)

type (
	ClientID   struct{ uuid.UUID }
	ClientType uint

	Client struct {
		ID      ClientID
		OwnerID shared.UserID

		Address        url.URL
		Type           ClientType
		Authentication authentication.Authentication
	}

	ReleaseDownload struct {
		ReleaseID       releases.ReleaseID
		Filepath        string
		Status          Status
		ProgressDecimal uint
		Hashes          []hashing.Hash
	}

	ReleaseCandidate struct {
		ReleaseID   releases.ReleaseID
		Tags        []string
		Category    string
		ShouldPause bool
		Filepath    string
		Name        string

		ContentType string
		Binary      []byte
	}
)

const (
	ClientTypeUnknown ClientType = iota
	ClientTypeQBittorrent
	clientTypeSentinel
)

var ClientTypeStringer = map[ClientType]string{
	ClientTypeQBittorrent: "qBittorrent",
}

func NewClient(
	ownerID shared.UserID,
	clientType ClientType,
	address url.URL,
	auth authentication.Authentication,
) (*Client, error) {
	return &Client{
		ID:             shared.NewID[ClientID](),
		OwnerID:        ownerID,
		Address:        address,
		Type:           clientType,
		Authentication: auth,
	}, nil
}

func (s ClientType) String() string {
	if value, ok := ClientTypeStringer[s]; ok {
		return value
	}
	return ""
}

func (s ClientType) IsValid() bool {
	return s > ClientTypeUnknown && s < clientTypeSentinel
}
