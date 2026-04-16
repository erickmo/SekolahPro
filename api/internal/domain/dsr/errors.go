package dsr

import "errors"

var (
	ErrNotFound          = errors.New("data subject request tidak ditemukan")
	ErrRequestTypeEmpty  = errors.New("request type tidak boleh kosong")
	ErrRequestorEmpty    = errors.New("requestor name tidak boleh kosong")
	ErrSubjectEmpty      = errors.New("subject id tidak boleh kosong")
	ErrInvalidStatus     = errors.New("status tidak valid")
	ErrAlreadyCompleted  = errors.New("request sudah selesai, tidak bisa diubah")
)
