package mappers

import (
	"encoding/json"

	"github.com/sonalys/animeman/internal/domain/authentication"
)

type (
	Authentication struct {
		Type string `json:"type,omitempty"`
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
	case "apiKey":
		return authentication.NewAPIKeyAuthentication(target.Key)
	default:
		return authentication.NewUserPasswordAuthentication(target.Username, target.Password)
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
	default:
		model = Authentication{
			Type: "userPassword",
			AuthenticationUserPassword: &AuthenticationUserPassword{
				Username: from.Username,
				Password: from.Password,
			},
		}
	}

	buf, _ := json.Marshal(model)
	return buf
}
