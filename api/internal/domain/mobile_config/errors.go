package mobile_config

import "errors"

var (
	ErrNotFound          = errors.New("konfigurasi mobile app tidak ditemukan")
	ErrDuplicateVariant  = errors.New("konfigurasi untuk varian ini sudah ada")
	ErrInvalidAppVariant = errors.New("app_variant tidak valid (pilih: student, parent, staff, full)")
	ErrInvalidPlatform   = errors.New("platform tidak valid (pilih: android, ios, both)")
	ErrInvalidVersion    = errors.New("format versi tidak valid (gunakan semver: x.y.z)")
	ErrEmptyAPIBaseURL   = errors.New("api_base_url tidak boleh kosong")
	ErrVersionTooLong    = errors.New("versi terlalu panjang (maks 20 karakter)")
	ErrAPIBaseURLTooLong = errors.New("api_base_url terlalu panjang (maks 255 karakter)")
)
