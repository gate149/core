package pkg

import (
	"github.com/google/uuid"
)

func ParseUUIDv4(s string) (uuid.UUID, error) {
	u, err := uuid.Parse(s)

	if err != nil {
		return uuid.Nil, err
	}

	if u.Version() != 4 || u.Variant() != uuid.RFC4122 {
		return uuid.Nil, Wrap(ErrBadInput, nil, "", "string is not a valid UUIDv4")
	}

	return u, nil
}
