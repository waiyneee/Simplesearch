package suggest

import (
	"context"
	"time"

	"github.com/redis/go-redis/v9"
)

// Cache is shared with corrector.go
type Cache interface {
	Get(key string) (val string, ok bool)
	Set(key, val string)
}

// --------------------
// Memory cache -->we dont want each and every lookup for redis
// --------------------
type MemoryCache struct {
	store map[string]string
}

func NewMemoryCache() *MemoryCache {
	return &MemoryCache{store: map[string]string{}}
}

func (m *MemoryCache) Get(key string) (string, bool) {
	val, ok := m.store[key]
	return val, ok
}

func (m *MemoryCache) Set(key, val string) {
	m.store[key] = val
}

// --------------------
// Redis cache
// --------------------
type RedisCache struct {
	client *redis.Client
	ttl    time.Duration
}

func NewRedisCache(redisURL string, ttl time.Duration) (*RedisCache, error) {
	opt, err := redis.ParseURL(redisURL)
	if err != nil {
		return nil, err
	}
	rdb := redis.NewClient(opt)
	return &RedisCache{client: rdb, ttl: ttl}, nil
}

func (r *RedisCache) Get(key string) (string, bool) {
	ctx := context.Background()
	val, err := r.client.Get(ctx, key).Result()
	if err == redis.Nil || err != nil {
		return "", false
	}
	return val, true
}

func (r *RedisCache) Set(key, val string) {
	ctx := context.Background()
	_ = r.client.Set(ctx, key, val, r.ttl).Err()
}
