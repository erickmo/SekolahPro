package consent

import "errors"

var (
	ErrNotFound         = errors.New("consent record not found")
	ErrAlreadyWithdrawn = errors.New("consent already withdrawn")
)
