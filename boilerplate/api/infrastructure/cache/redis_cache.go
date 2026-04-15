package cache

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// RedisCache mengimplementasi Cache menggunakan Redis.
// Semua key di-prefix dengan namespace untuk menghindari collision antar service.
type RedisCache struct {
	client    *redis.Client
	namespace string
}

// NewRedisCache membuat instance baru RedisCache.
// namespace: prefix yang ditambahkan ke semua key, e.g. "boilerplate".
func NewRedisCache(client *redis.Client, namespace string) *RedisCache {
	return &RedisCache{client: client, namespace: namespace}
}

// Get mengambil value berdasarkan key.
// Mengembalikan ErrCacheMiss jika key tidak ada atau sudah expired.
func (c *RedisCache) Get(ctx context.Context, key string) ([]byte, error) {
	val, err := c.client.Get(ctx, c.prefixed(key)).Bytes()
	if err == redis.Nil {
		return nil, ErrCacheMiss
	}
	if err != nil {
		return nil, err
	}
	return val, nil
}

// Set menyimpan value dengan TTL. TTL 0 berarti tidak ada expiry.
func (c *RedisCache) Set(ctx context.Context, key string, value []byte, ttl time.Duration) error {
	return c.client.Set(ctx, c.prefixed(key), value, ttl).Err()
}

// Delete menghapus satu atau lebih key. Error diabaikan jika key tidak ada.
func (c *RedisCache) Delete(ctx context.Context, keys ...string) error {
	prefixed := make([]string, len(keys))
	for i, k := range keys {
		prefixed[i] = c.prefixed(k)
	}
	return c.client.Del(ctx, prefixed...).Err()
}

func (c *RedisCache) prefixed(key string) string {
	return c.namespace + ":" + key
}
