// Package cache menyediakan abstraksi caching dan implementasi Redis.
// Didesain untuk digunakan sebagai decorator di atas repository layer.
package cache

import (
	"context"
	"errors"
	"time"
)

// ErrCacheMiss dikembalikan ketika key tidak ditemukan di cache.
var ErrCacheMiss = errors.New("cache miss")

// Cache mendefinisikan kontrak minimal untuk operasi caching.
type Cache interface {
	// Get mengambil value berdasarkan key. Mengembalikan ErrCacheMiss jika tidak ada.
	Get(ctx context.Context, key string) ([]byte, error)
	// Set menyimpan value dengan TTL tertentu. TTL 0 berarti tidak ada expiry.
	Set(ctx context.Context, key string, value []byte, ttl time.Duration) error
	// Delete menghapus satu atau lebih key dari cache.
	Delete(ctx context.Context, keys ...string) error
}
