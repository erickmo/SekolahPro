package lifecycle

import "errors"

var (
	ErrNotFound            = errors.New("membership lifecycle tidak ditemukan")
	ErrInvalidTransition   = errors.New("transisi status tidak valid")
	ErrAlreadyActive       = errors.New("anggota sudah aktif")
	ErrAlreadySuspended    = errors.New("anggota sudah ditangguhkan")
	ErrCannotReactivate    = errors.New("anggota tidak bisa diaktifkan kembali")
	ErrNasabahIDEmpty      = errors.New("nasabah id tidak boleh kosong")
	ErrMembershipNoEmpty   = errors.New("membership number tidak boleh kosong")
)
