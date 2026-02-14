package shared

type StringError string

func (e StringError) Error() string {
	return string(e)
}
