package limiter

import "time"

type RateLimiter interface {
	IsAllowed(key string, limit int, limitDuration, blockDuration time.Duration) bool
}
