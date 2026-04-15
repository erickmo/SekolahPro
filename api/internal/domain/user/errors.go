package user

import "errors"

var (
	ErrNotFound       = errors.New("user tidak ditemukan")
	ErrEmailEmpty     = errors.New("email wajib diisi")
	ErrEmailInvalid   = errors.New("format email tidak valid")
	ErrEmailExists    = errors.New("email sudah digunakan")
	ErrFullNameEmpty  = errors.New("full_name wajib diisi")
	ErrPasswordEmpty  = errors.New("password wajib diisi")
	ErrPasswordWeak   = errors.New("password minimal 8 karakter")
	ErrInactive       = errors.New("akun tidak aktif")
	ErrWrongPassword  = errors.New("email atau password salah")
)
