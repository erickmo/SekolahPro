package incident

import "errors"

var (
	ErrNotFound          = errors.New("incident tidak ditemukan")
	ErrTitleEmpty        = errors.New("title tidak boleh kosong")
	ErrTitleTooLong      = errors.New("title terlalu panjang (maks 255 karakter)")
	ErrSeverityEmpty     = errors.New("severity tidak boleh kosong")
	ErrInvalidSeverity   = errors.New("severity tidak valid (harus: critical, high, medium, low)")
	ErrInvalidStatus     = errors.New("status tidak valid (harus: open, investigating, resolved, closed)")
	ErrAlreadyResolved   = errors.New("incident sudah di-resolve")
)
