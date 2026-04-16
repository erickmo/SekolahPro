package biometric

import "errors"

var (
	ErrNotFound           = errors.New("biometric enrollment tidak ditemukan")
	ErrAlreadyEnrolled    = errors.New("nasabah sudah terdaftar untuk tipe biometric ini")
	ErrMaxAttempts        = errors.New("percobaan verifikasi melebihi batas maksimum")
	ErrVerificationFailed = errors.New("verifikasi biometric gagal")
	ErrInvalidBiometricType = errors.New("tipe biometric tidak valid")
	ErrTemplateHashEmpty  = errors.New("template hash tidak boleh kosong")
	ErrNasabahEmpty       = errors.New("nasabah id tidak boleh kosong")
	ErrNotActive          = errors.New("biometric enrollment tidak aktif")
)
