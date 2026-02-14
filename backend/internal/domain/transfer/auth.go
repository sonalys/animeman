package transfer

type (
	AuthenticationType uint

	Authentication struct {
		Type AuthenticationType

		*AuthenticationUserPassword
	}

	AuthenticationUserPassword struct {
		Username string
		Password []byte
	}
)

const (
	AuthenticationTypeUnknown AuthenticationType = iota
	AuthenticationTypeUserPassword
	authenticationTypeSentinel
)

func (s AuthenticationType) String() string {
	switch s {
	case AuthenticationTypeUserPassword:
		return "userPassword"
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
