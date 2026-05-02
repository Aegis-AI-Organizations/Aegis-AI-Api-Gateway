package middleware

import (
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"golang.org/x/time/rate"
)

type IPRateLimiter struct {
	ips map[string]*rate.Limiter
	mu  *sync.RWMutex
	r   rate.Limit
	b   int
}

func NewIPRateLimiter(r rate.Limit, b int) *IPRateLimiter {
	return &IPRateLimiter{
		ips: make(map[string]*rate.Limiter),
		mu:  &sync.RWMutex{},
		r:   r,
		b:   b,
	}
}

func (i *IPRateLimiter) GetLimiter(ip string) *rate.Limiter {
	i.mu.RLock()
	limiter, exists := i.ips[ip]
	i.mu.RUnlock()

	if !exists {
		i.mu.Lock()
		defer i.mu.Unlock()
		// Double check
		limiter, exists = i.ips[ip]
		if !exists {
			limiter = rate.NewLimiter(i.r, i.b)
			i.ips[ip] = limiter
		}
	}

	return limiter
}

func RateLimiter(limit rate.Limit, burst int) gin.HandlerFunc {
	limiter := NewIPRateLimiter(limit, burst)

	return func(c *gin.Context) {
		ip := c.ClientIP()
		if !limiter.GetLimiter(ip).Allow() {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}
		c.Next()
	}
}

// RedisRateLimiter implements a distributed rate limiter using Redis.
// It uses a fixed-window counter with a 1-minute expiration.
func RedisRateLimiter(rdb interface{}) gin.HandlerFunc {
	// We use a local interface to access the redis.Client without a circular dependency on the db package.
	type redisWrapper interface {
		GetClient() *redis.Client
	}

	// Use reflection to check if the underlying value is nil (interface containing a nil pointer)
	if rdb == nil || (reflect.ValueOf(rdb).Kind() == reflect.Ptr && reflect.ValueOf(rdb).IsNil()) {
		return func(c *gin.Context) { c.Next() }
	}

	wrapper, ok := rdb.(redisWrapper)
	if !ok {
		// Fallback or ignore if Redis is not available
		return func(c *gin.Context) { c.Next() }
	}

	client := wrapper.GetClient()
	if client == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()
		ip := c.ClientIP()
		key := "ratelimit:" + ip

		// Atomic increment
		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("Redis ratelimit error: %v", err)
			c.Next()
			return
		}

		// Set expiration on first hit
		if count == 1 {
			client.Expire(ctx, key, time.Minute)
		}

		// Limit: 100 requests per minute
		if count > 100 {
			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}
