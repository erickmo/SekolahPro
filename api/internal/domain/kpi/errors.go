package kpi

import "errors"

var (
	ErrDefinitionNotFound    = errors.New("kpi definition tidak ditemukan")
	ErrMeasurementNotFound   = errors.New("kpi measurement tidak ditemukan")
	ErrScenarioNotFound      = errors.New("stress test scenario tidak ditemukan")
	ErrNameEmpty             = errors.New("name tidak boleh kosong")
	ErrCodeEmpty             = errors.New("code tidak boleh kosong")
	ErrInvalidCategory       = errors.New("kategori KPI tidak valid")
	ErrInvalidFrequency      = errors.New("frekuensi tidak valid")
	ErrValueOutOfRange       = errors.New("measured value di luar jangkauan")
	ErrScenarioRunning       = errors.New("scenario sedang berjalan, tidak bisa diubah")
)
