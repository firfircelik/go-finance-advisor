package middleware

import (
	"net/http"
	"sync"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimiter implements IP-based rate limiting
type RateLimiter struct {
	limiters map[string]*rate.Limiter
	mu       sync.RWMutex
	rps      rate.Limit // requests per second
	burst    int        // max burst size
}

// NewRateLimiter creates a new rate limiter
// rps: requests per second allowed
// burst: maximum burst size (tokens in bucket)
func NewRateLimiter(rps int, burst int) *RateLimiter {
	return &RateLimiter{
		limiters: make(map[string]*rate.Limiter),
		rps:      rate.Limit(rps),
		burst:    burst,
	}
}

// getLimiter returns the rate limiter for a given IP
func (rl *RateLimiter) getLimiter(ip string) *rate.Limiter {
	rl.mu.RLock()
	limiter, exists := rl.limiters[ip]
	rl.mu.RUnlock()

	if exists {
		return limiter
	}

	// Create new limiter for this IP
	rl.mu.Lock()
	limiter = rate.NewLimiter(rl.rps, rl.burst)
	rl.limiters[ip] = limiter
	rl.mu.Unlock()

	return limiter
}

// RateLimitMiddleware returns a Gin middleware that implements rate limiting
func (rl *RateLimiter) RateLimitMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		ip := c.ClientIP()
		limiter := rl.getLimiter(ip)

		if !limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":        "Rate limit exceeded",
				"retry_after":  "Please try again later",
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// CleanupLimiters periodically removes old limiters to prevent memory leaks
// Call this in a goroutine: go limiter.CleanupLimiters(5 * time.Minute)
func (rl *RateLimiter) CleanupLimiters(interval int) {
	// For now, this is a simple implementation
	// In production, you'd want to track last access time and remove stale entries
	rl.mu.Lock()
	defer rl.mu.Unlock()

	// If we have too many limiters (> 10000), clear old ones
	if len(rl.limiters) > 10000 {
		rl.limiters = make(map[string]*rate.Limiter)
	}
}
