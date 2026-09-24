package middleware

import (
	"net/http"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/gofiber/fiber/v2"
)

type rateLimitBucket struct {
	count     int
	resetTime time.Time
}

type rateLimitConfig struct {
	generalRequestsPerMinute int
	authRequestsPerMinute    int
	otpRequestsPerWindow     int
	generalWindow            time.Duration
	authWindow               time.Duration
	otpWindow                time.Duration
}

var (
	rateLimitStore = make(map[string]*rateLimitBucket)
	mu             sync.RWMutex
)

const cleanupInterval = 5 * time.Minute

var config = loadRateLimitConfig()

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

func loadRateLimitConfig() rateLimitConfig {
	return rateLimitConfig{
		generalRequestsPerMinute: envInt("RATE_LIMIT_REQUESTS_PER_MINUTE", 120),
		authRequestsPerMinute:    envInt("RATE_LIMIT_AUTH_REQUESTS_PER_MINUTE", 10),
		otpRequestsPerWindow:     envInt("RATE_LIMIT_OTP_REQUESTS", 5),
		generalWindow:            time.Minute,
		authWindow:               time.Minute,
		otpWindow:                10 * time.Minute,
	}
}

func envInt(name string, fallback int) int {
	value, err := strconv.Atoi(os.Getenv(name))
	if err != nil || value <= 0 {
		return fallback
	}
	return value
}

func RateLimiter(c *fiber.Ctx) error {
	if isRateLimitExempt(c.Path()) {
		return c.Next()
	}

	ip := c.IP()
	bucketName, limit, window, exempt := rateLimitRule(c.Path())
	if exempt {
		return c.Next()
	}

	mu.Lock()
	now := time.Now()
	key := rateLimitKey(ip, bucketName)
	bucket, exists := rateLimitStore[key]

	if !exists || now.After(bucket.resetTime) {
		rateLimitStore[key] = &rateLimitBucket{
			count:     1,
			resetTime: now.Add(window),
		}
		mu.Unlock()
		return c.Next()
	}

	bucket.count++
	if bucket.count > limit {
		retryAfter := max(1, int(time.Until(bucket.resetTime).Seconds()+0.999))
		mu.Unlock()
		c.Set("Retry-After", strconv.Itoa(retryAfter))
		return c.Status(http.StatusTooManyRequests).JSON(fiber.Map{
			"success": false,
			"message": "Rate limit exceeded",
		})
	}
	mu.Unlock()

	return c.Next()
}

func rateLimitKey(ip, bucketName string) string {
	return bucketName + ":" + ip
}

func isAuthPath(path string) bool {
	switch path {
	case "/api/login", "/api/register", "/api/verify-otp", "/api/resend-otp", "/api/auth/google", "/api/auth/google/username":
		return true
	default:
		return false
	}
}

func rateLimitRule(path string) (string, int, time.Duration, bool) {
	if isRateLimitExempt(path) {
		return "", 0, 0, true
	}
	if path == "/api/verify-otp" || path == "/api/resend-otp" {
		return "otp", config.otpRequestsPerWindow, config.otpWindow, false
	}
	if isAuthPath(path) {
		return "auth", config.authRequestsPerMinute, config.authWindow, false
	}
	return "general", config.generalRequestsPerMinute, config.generalWindow, false
}

func isRateLimitExempt(path string) bool {
	return path == "/docs" || strings.HasPrefix(path, "/docs/") ||
		path == "/uploads" || strings.HasPrefix(path, "/uploads/")
}
