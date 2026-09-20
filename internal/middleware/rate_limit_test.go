package middleware

import (
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func resetRateLimitTestState(t *testing.T, generalLimit, authLimit int) {
	t.Helper()

	mu.Lock()
	previousStore := rateLimitStore
	previousConfig := config
	rateLimitStore = make(map[string]*rateLimitBucket)
	config = rateLimitConfig{
		generalRequestsPerMinute: generalLimit,
		authRequestsPerMinute:    authLimit,
		otpRequestsPerWindow:     2,
		generalWindow:            time.Minute,
		authWindow:               time.Minute,
		otpWindow:                time.Minute,
	}
	mu.Unlock()

	t.Cleanup(func() {
		mu.Lock()
		rateLimitStore = previousStore
		config = previousConfig
		mu.Unlock()
	})
}

func TestRateLimiterUsesSeparateGeneralAndAuthBuckets(t *testing.T) {
	resetRateLimitTestState(t, 2, 1)

	app := fiber.New()
	app.Use(RateLimiter)
	app.All("/*", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	for range 2 {
		response, err := app.Test(httptest.NewRequest("GET", "/api/data", nil))
		if err != nil || response.StatusCode != fiber.StatusNoContent {
			t.Fatalf("general request failed: status=%d err=%v", response.StatusCode, err)
		}
	}

	response, err := app.Test(httptest.NewRequest("POST", "/api/login", nil))
	if err != nil || response.StatusCode != fiber.StatusNoContent {
		t.Fatalf("auth bucket should be independent: status=%d err=%v", response.StatusCode, err)
	}

	response, err = app.Test(httptest.NewRequest("POST", "/api/login", nil))
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusTooManyRequests || response.Header.Get("Retry-After") == "" {
		t.Fatalf("expected 429 with Retry-After, got status=%d header=%q", response.StatusCode, response.Header.Get("Retry-After"))
	}
}

func TestRateLimiterDoesNotHoldLockWhileHandlerRuns(t *testing.T) {
	resetRateLimitTestState(t, 10, 10)

	app := fiber.New()
	app.Use(RateLimiter)
	entered := make(chan struct{}, 2)
	release := make(chan struct{}, 2)
	app.Get("/api/data", func(c *fiber.Ctx) error {
		entered <- struct{}{}
		<-release
		return c.SendStatus(fiber.StatusNoContent)
	})

	var waitGroup sync.WaitGroup
	waitGroup.Add(2)
	for range 2 {
		go func() {
			defer waitGroup.Done()
			_, _ = app.Test(httptest.NewRequest("GET", "/api/data", nil))
		}()
	}

	for range 2 {
		select {
		case <-entered:
		case <-time.After(200 * time.Millisecond):
			t.Fatal("second request could not enter handler while first request was running")
		}
	}

	release <- struct{}{}
	release <- struct{}{}
	waitGroup.Wait()
}

func TestRateLimiterExemptsStaticPaths(t *testing.T) {
	resetRateLimitTestState(t, 1, 1)

	app := fiber.New()
	app.Use(RateLimiter)
	app.Get("/docs/index.html", func(c *fiber.Ctx) error {
		return c.SendStatus(fiber.StatusNoContent)
	})

	for range 2 {
		response, err := app.Test(httptest.NewRequest("GET", "/docs/index.html", nil))
		if err != nil || response.StatusCode != fiber.StatusNoContent {
			t.Fatalf("static path should be exempt: status=%d err=%v", response.StatusCode, err)
		}
	}
}
