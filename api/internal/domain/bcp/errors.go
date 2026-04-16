package bcp

import "errors"

var (
	ErrNotFound          = errors.New("backup strategy tidak ditemukan")
	ErrStrategyTypeEmpty = errors.New("strategy type tidak boleh kosong")
	ErrFrequencyEmpty    = errors.New("frequency tidak boleh kosong")
	ErrFrequencyTooLong  = errors.New("frequency terlalu panjang (maks 100 karakter)")
	ErrStorageTooLong    = errors.New("storage location terlalu panjang (maks 500 karakter)")
	ErrInvalidStrategy   = errors.New("strategy type tidak valid (harus: full, incremental, differential, snapshot)")
)
