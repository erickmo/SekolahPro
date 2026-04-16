package education

import "errors"

var (
	ErrNotFound          = errors.New("kursus pendidikan tidak ditemukan")
	ErrTitleEmpty        = errors.New("judul kursus tidak boleh kosong")
	ErrTitleTooLong      = errors.New("judul kursus terlalu panjang (maks 255 karakter)")
	ErrInvalidCourseType = errors.New("jenis kursus tidak valid (mandatory, elective, certification)")
	ErrInvalidScoreRange = errors.New("nilai kelulusan harus antara 0-100")
	ErrInvalidAttempts   = errors.New("maksimal percobaan harus lebih dari 0")
)
