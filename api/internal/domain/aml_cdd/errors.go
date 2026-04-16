package aml_cdd

import "errors"

var (
	ErrNotFound       = errors.New("cdd record not found")
	ErrAlreadyExists  = errors.New("cdd record already exists for this nasabah")
)
