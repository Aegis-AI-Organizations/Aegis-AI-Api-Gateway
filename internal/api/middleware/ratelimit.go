package middleware

import (
	"log"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/Aegis-AI-Organizations/aegis-ai-api-gateway/internal/types"
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
			ttl, err := client.TTL(ctx, key).Result()
			if err == nil && ttl <= 0 {
				client.Expire(ctx, key, time.Minute)
			} else if err != nil {
				log.Printf("Redis ratelimit TTL error: %v", err)
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Too many requests. Please try again later.",
			})
			return
		}

		c.Next()
	}
}

// AgentRateLimiter implements a distributed rate limiter for Agents using Redis.
// It limits to 50 requests per second per Agent (tenant_id).
func AgentRateLimiter(rdb interface{}) gin.HandlerFunc {
	type redisWrapper interface {
		GetClient() *redis.Client
	}

	if rdb == nil || (reflect.ValueOf(rdb).Kind() == reflect.Ptr && reflect.ValueOf(rdb).IsNil()) {
		return func(c *gin.Context) { c.Next() }
	}

	wrapper, ok := rdb.(redisWrapper)
	if !ok {
		return func(c *gin.Context) { c.Next() }
	}

	client := wrapper.GetClient()
	if client == nil {
		return func(c *gin.Context) { c.Next() }
	}

	return func(c *gin.Context) {
		ctx := c.Request.Context()

		// Get tenant_id from context (set by AgentAuthMiddleware)
		tenantID, exists := c.Get(string(types.AgentTenantIDKey))
		if !exists {
			// If not an agent request, skip rate limiting or fall back to IP?
			// For now, skip to avoid blocking other routes if applied incorrectly.
			c.Next()
			return
		}

		key := "ratelimit:agent:" + tenantID.(string)
		if agentID, exists := c.Get(string(types.AgentIDKey)); exists {
			if agentIDString, ok := agentID.(string); ok && agentIDString != "" {
				key += ":" + agentIDString
			}
		}

		// Atomic increment
		count, err := client.Incr(ctx, key).Result()
		if err != nil {
			log.Printf("Redis agent ratelimit error: %v", err)
			c.Next()
			return
		}

		// Set expiration to 1 second for "per second" limiting
		if count == 1 {
			client.Expire(ctx, key, time.Second)
		}

		// Limit: 50 requests per second
		if count > 50 {
			ttl, err := client.TTL(ctx, key).Result()
			if err == nil && ttl <= 0 {
				client.Expire(ctx, key, time.Second)
			} else if err != nil {
				log.Printf("Redis agent ratelimit TTL error: %v", err)
			}

			c.AbortWithStatusJSON(http.StatusTooManyRequests, gin.H{
				"error": "Agent request limit exceeded (50 req/sec).",
			})
			return
		}

		c.Next()
	}
}
