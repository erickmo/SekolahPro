package governance

import "errors"

var (
	ErrNotFound      = errors.New("governance position not found")
	ErrAlreadyActive = errors.New("position already has an active holder")
)
