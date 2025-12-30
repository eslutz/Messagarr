package resilience

import (
	"sync"
	"time"
)

// RateLimiter implements token bucket rate limiting per channel
type RateLimiter struct {
	mu       sync.Mutex
	buckets  map[string]*tokenBucket
	capacity int
	refill   int
}

// tokenBucket represents a token bucket for a channel
type tokenBucket struct {
	tokens     int
	lastRefill time.Time
}

// NewRateLimiter creates a new rate limiter
func NewRateLimiter(capacity, refillRate int) *RateLimiter {
	return &RateLimiter{
		buckets:  make(map[string]*tokenBucket),
		capacity: capacity,
		refill:   refillRate,
	}
}

// Wait waits if rate limit is exceeded for a channel
func (rl *RateLimiter) Wait(channel string) {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	bucket, exists := rl.buckets[channel]
	if !exists {
		bucket = &tokenBucket{
			tokens:     rl.capacity,
			lastRefill: time.Now(),
		}
		rl.buckets[channel] = bucket
	}

	// Refill tokens based on time elapsed
	now := time.Now()
	elapsed := now.Sub(bucket.lastRefill)
	tokensToAdd := int(elapsed.Seconds()) * rl.refill
	if tokensToAdd > 0 {
		bucket.tokens += tokensToAdd
		if bucket.tokens > rl.capacity {
			bucket.tokens = rl.capacity
		}
		bucket.lastRefill = now
	}

	// Wait if no tokens available
	for bucket.tokens <= 0 {
		rl.mu.Unlock()
		time.Sleep(100 * time.Millisecond)
		rl.mu.Lock()

		// Refill again after waiting
		now = time.Now()
		elapsed = now.Sub(bucket.lastRefill)
		tokensToAdd = int(elapsed.Seconds()) * rl.refill
		if tokensToAdd > 0 {
			bucket.tokens += tokensToAdd
			if bucket.tokens > rl.capacity {
				bucket.tokens = rl.capacity
			}
			bucket.lastRefill = now
		}
	}

	// Consume a token
	bucket.tokens--
}
