package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

type RateLimiter struct {
	requests map[string][]time.Time
	mu       sync.Mutex
}

func NewRateLimiter() *RateLimiter {
	return &RateLimiter{
		requests: make(map[string][]time.Time),
	}
}

func (rl *RateLimiter) RateLimit() gin.HandlerFunc {
	return func(c *gin.Context) {
		rl.mu.Lock()
		defer rl.mu.Unlock()

		ip := c.ClientIP()
		now := time.Now()

		// Clean old requests
		if old, exists := rl.requests[ip]; exists {
			var valid []time.Time
			for _, t := range old {
				if now.Sub(t) < time.Minute {
					valid = append(valid, t)
				}
			}
			rl.requests[ip] = valid
		}

		// Check rate limit (100 requests per minute)
		if len(rl.requests[ip]) >= 100 {
			c.JSON(429, gin.H{"error": "Too many requests"})
			c.Abort()
			return
		}

		rl.requests[ip] = append(rl.requests[ip], now)
		c.Next()
	}
}
