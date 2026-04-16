package collection

import "errors"

var (
	ErrCaseNotFound          = errors.New("collection case tidak ditemukan")
	ErrCaseNumberEmpty       = errors.New("nomor case tidak boleh kosong")
	ErrInvalidAgingBucket    = errors.New("aging bucket tidak valid")
	ErrInvalidCaseStatus     = errors.New("status case tidak valid")
	ErrInvalidResolutionType = errors.New("tipe resolusi tidak valid")
	ErrCaseAlreadyResolved   = errors.New("case sudah diresolusi")
	ErrNegativeDPD           = errors.New("DPD tidak boleh negatif")
	ErrNegativeOverdue       = errors.New("jumlah overdue tidak boleh negatif")
	ErrCollectorNotAssigned  = errors.New("collector belum ditugaskan")
	ErrActivityNotFound      = errors.New("aktivitas collection tidak ditemukan")
	ErrPerformanceNotFound   = errors.New("performa collector tidak ditemukan")
	ErrInvalidPeriodMonth    = errors.New("format period_month tidak valid (contoh: 2026-01)")
	ErrInvalidActivityType   = errors.New("tipe aktivitas tidak valid")
)
