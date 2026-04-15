package vernon

import "errors"

var (
	// ErrNotFound dikembalikan ketika entitas tidak ditemukan atau sudah dihapus.
	ErrNotFound = errors.New("entity not found")

	// ErrVersionConflict dikembalikan ketika CAS update gagal karena versi berbeda.
	// Caller harus re-fetch dan retry.
	ErrVersionConflict = errors.New("sync version conflict — re-fetch and retry")

	// ErrCycleDetected dikembalikan saat startup jika ada circular autoload dependency.
	ErrCycleDetected = errors.New("circular autoload dependency detected")
)
