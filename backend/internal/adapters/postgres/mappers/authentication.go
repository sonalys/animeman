package mappers

import (
	"encoding/json"

	"github.com/sonalys/animeman/internal/adapters/postgres/sqlcgen"
	"github.com/sonalys/animeman/internal/domain/authentication"
)

type (
	Authentication struct {
		Type sqlcgen.AuthType `json:"type,omitempty"`
		*AuthenticationUserPassword
		*AuthenticationAPIKey
	}

	AuthenticationUserPassword struct {
		Username string `json:"username,omitempty"`
		Password []byte `json:"password,omitempty"`
	}

	AuthenticationAPIKey struct {
		Key string `json:"key,omitempty"`
	}
)

func NewAuthentication(from []byte) authentication.Authentication {
	var target Authentication

	_ = json.Unmarshal(from, &target)

	switch target.Type {
	case sqlcgen.AuthTypeApiKey:
		return authentication.NewAPIKeyAuthentication(target.Key)
	case sqlcgen.AuthTypeUserPassword:
		return authentication.NewUserPasswordAuthentication(target.Username, target.Password)
	default:
		return authentication.Authentication{}
	}
}

func NewAuthenticationModel(from authentication.Authentication) []byte {
	var model Authentication

	switch from.Type {
	case authentication.AuthenticationTypeAPIKey:
		model = Authentication{
			Type: "apiKey",
			AuthenticationAPIKey: &AuthenticationAPIKey{
				Key: from.Key,
			},
		}
	case authentication.AuthenticationTypeUserPassword:
		model = Authentication{
			Type: sqlcgen.AuthTypeUserPassword,
			AuthenticationUserPassword: &AuthenticationUserPassword{
				Username: from.Username,
				Password: from.Password,
			},
		}
	case authentication.AuthenticationTypeNone:
		model = Authentication{
			Type: "none",
		}
	}

	buf, _ := json.Marshal(model)
	return buf
}

func NewAuthenticationTypeModel(from authentication.AuthenticationType) sqlcgen.AuthType {
	switch from {
	case authentication.AuthenticationTypeAPIKey:
		return sqlcgen.AuthTypeApiKey
	case authentication.AuthenticationTypeUserPassword:
		return sqlcgen.AuthTypeUserPassword
	default:
		return sqlcgen.AuthTypeNone
	}
}
