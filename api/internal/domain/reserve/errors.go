package reserve

import "errors"

var (
	ErrFundNotFound          = errors.New("reserve fund tidak ditemukan")
	ErrTransactionNotFound   = errors.New("reserve fund transaction tidak ditemukan")
	ErrNameEmpty             = errors.New("name tidak boleh kosong")
	ErrFundTypeEmpty         = errors.New("fund type tidak boleh kosong")
	ErrInvalidAmount         = errors.New("amount tidak valid")
	ErrInsufficientBalance   = errors.New("saldo dana cadangan tidak mencukupi")
	ErrFundFrozen            = errors.New("dana cadangan dibekukan, tidak bisa transaksi")
	ErrDuplicateReference    = errors.New("nomor referensi sudah digunakan")
)
