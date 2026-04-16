package biometric_log

import "errors"

var (
	ErrNotFound          = errors.New("log verifikasi biometric tidak ditemukan")
	ErrEnrollmentEmpty   = errors.New("enrollment id tidak boleh kosong")
	ErrNasabahEmpty      = errors.New("nasabah id tidak boleh kosong")
	ErrResultEmpty       = errors.New("verification result tidak boleh kosong")
)
