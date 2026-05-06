package app

// Config carries optional runtime settings (cache, redis, etc.).
// although it pulls up network operations overhead but we are good with that
// as of now
type Config struct {
	CacheMode string // "memory" or "redis"
	RedisURL  string // e.g. redis://localhost:6379
}
