package middleware

import (
	"net/http"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type rateLimitBucket struct {
	count     int
	resetTime time.Time
}

var (
	rateLimitStore = make(map[string]*rateLimitBucket)
	mu             sync.RWMutex
)

const (
	requestsPerMinute = 60
	cleanupInterval   = 5 * time.Minute
)

func init() {
	go cleanupRateLimitStore()
}

func cleanupRateLimitStore() {
	ticker := time.NewTicker(cleanupInterval)
	defer ticker.Stop()
	for range ticker.C {
		mu.Lock()
		now := time.Now()
		for ip, bucket := range rateLimitStore {
			if now.After(bucket.resetTime) {
				delete(rateLimitStore, ip)
			}
		}
		mu.Unlock()
	}
}

func RateLimiter(c *fiber.Ctx) error {
	ip := c.IP()

	mu.Lock()
	defer mu.Unlock()

	now := time.Now()
	bucket, exists := rateLimitStore[ip]

	if !exists || now.After(bucket.resetTime) {
		rateLimitStore[ip] = &rateLimitBucket{
			count:     1,
			resetTime: now.Add(time.Minute),
		}
		return c.Next()
	}

	bucket.count++
	if bucket.count > requestsPerMinute {
		return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"message": "Rate limit exceeded",
		})
	}

	return c.Next()
}
