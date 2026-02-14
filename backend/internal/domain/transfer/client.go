package transfer

import (
	"net/url"

	"github.com/gofrs/uuid/v5"
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
		Authentication Authentication
	}

	ReleaseDownload struct {
		ReleaseID       releases.ReleaseID
		SavePath        string
		Status          Status
		ProgressDecimal uint
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

func (s ClientType) String() string {
	if value, ok := ClientTypeStringer[s]; ok {
		return value
	}
	return ""
}

func (s ClientType) IsValid() bool {
	return s > ClientTypeUnknown && s < clientTypeSentinel
}
