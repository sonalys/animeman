package shared

import "github.com/gofrs/uuid/v5"

type (
	ID     = struct{ uuid.UUID }
	UserID struct{ uuid.UUID }
)

func NewID[T ~struct{ uuid.UUID }]() T {
	return T{uuid.Must(uuid.NewV7())}
}

func ParseStringID[T ~struct{ uuid.UUID }](input string) T {
	return T{uuid.FromStringOrNil(input)}
}
