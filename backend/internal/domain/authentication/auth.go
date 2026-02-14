package authentication

type (
	AuthenticationType uint

	Authentication struct {
		Type AuthenticationType

		*AuthenticationUserPassword
		*AuthenticationAPIKey
	}

	AuthenticationUserPassword struct {
		Username string
		Password []byte
	}

	AuthenticationAPIKey struct {
		Key string
	}
)

const (
	AuthenticationTypeUnknown AuthenticationType = iota
	AuthenticationTypeUserPassword
	AuthenticationTypeAPIKey
	authenticationTypeSentinel
)

func (s AuthenticationType) String() string {
	switch s {
	case AuthenticationTypeUserPassword:
		return "userPassword"
	case AuthenticationTypeAPIKey:
		return "apikey"
	default:
		return "unknown"
	}
}

func (s AuthenticationType) IsValid() bool {
	return s > AuthenticationTypeUnknown && s < authenticationTypeSentinel
}

func NewUserPasswordAuthentication(
	username string,
	password []byte,
) Authentication {
	return Authentication{
		Type: AuthenticationTypeUserPassword,
		AuthenticationUserPassword: &AuthenticationUserPassword{
			Username: username,
			Password: password,
		},
	}
}

func NewAPIKeyAuthentication(key string) Authentication {
	return Authentication{
		Type: AuthenticationTypeAPIKey,
		AuthenticationAPIKey: &AuthenticationAPIKey{
			Key: key,
		},
	}
}
