package domain

type StringError string

const (
	ErrUniqueUsername        StringError = "username must be unique"
	ErrInvalidPasswordLength StringError = "password must be between 8 and 72 digits"
)

func (e StringError) Error() string {
	return string(e)
}
