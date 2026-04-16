package aml_rule

import "errors"

var (
	ErrNotFound     = errors.New("monitoring rule not found")
	ErrDuplicateCode = errors.New("rule code already exists")
)
