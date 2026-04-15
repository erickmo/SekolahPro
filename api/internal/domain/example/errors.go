package example

import "errors"

var (
	ErrNotFound    = errors.New("example tidak ditemukan")
	ErrNameEmpty   = errors.New("nama tidak boleh kosong")
	ErrNameTooLong = errors.New("nama terlalu panjang (maks 255 karakter)")
)
