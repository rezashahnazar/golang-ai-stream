package middleware

import (
	"net/http"
	"sync"
	"time"

	"golang-ai-stream/errors"
)

type RateLimiter struct {
	tokens         float64
	capacity      float64
	refillRate    float64
	lastTimestamp time.Time
	mu            sync.Mutex
}

func NewRateLimiter(requestsPerSecond float64) *RateLimiter {
	return &RateLimiter{
		tokens:         requestsPerSecond,
		capacity:      requestsPerSecond,
		refillRate:    requestsPerSecond,
		lastTimestamp: time.Now(),
	}
}

func (rl *RateLimiter) tryConsume() bool {
	rl.mu.Lock()
	defer rl.mu.Unlock()

	now := time.Now()
	timePassed := now.Sub(rl.lastTimestamp).Seconds()
	rl.tokens = min(rl.capacity, rl.tokens+(timePassed*rl.refillRate))
	rl.lastTimestamp = now

	if rl.tokens >= 1 {
		rl.tokens--
		return true
	}
	return false
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func RateLimit(limiter *RateLimiter) func(http.Handler) http.Handler {
	return func(next http.Handler) http.Handler {
		return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if !limiter.tryConsume() {
				errors.ErrTooManyRequests("Rate limit exceeded").RespondWithError(w)
				return
			}
			next.ServeHTTP(w, r)
		})
	}
} 