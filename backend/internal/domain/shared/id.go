package shared

import "github.com/gofrs/uuid/v5"

type (
	UserID struct{ uuid.UUID }
)

func NewID[T ~struct{ uuid.UUID }]() T {
	return T{uuid.Must(uuid.NewV7())}
}

func ParseID[T ~struct{ uuid.UUID }](input string) T {
	return T{uuid.FromStringOrNil(input)}
}
