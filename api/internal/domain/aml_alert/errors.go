package aml_alert

import "errors"

var (
	ErrNotFound         = errors.New("transaction alert not found")
	ErrAlreadyResolved  = errors.New("alert already resolved")
)
