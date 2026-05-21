package optional

import (
	"errors"
)

// ErrNotSet is returned when calling [Optional.Get] on an empty Optional.
var ErrNotSet = errors.New("optional value not set")

