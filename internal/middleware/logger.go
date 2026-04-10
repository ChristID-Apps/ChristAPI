package middleware

import (
	"fmt"
	"time"

	"github.com/gofiber/fiber/v2"
)

// ANSI color
const (
	Reset  = "\033[0m"
	Red    = "\033[31m"
	Green  = "\033[32m"
	Yellow = "\033[33m"
	Blue   = "\033[34m"
	Cyan   = "\033[36m"
	Gray   = "\033[90m"
)

func getStatusColor(status int) string {
	switch {
	case status >= 500:
		return Red
	case status >= 400:
		return Yellow
	case status >= 300:
		return Cyan
	default:
		return Green
	}
}

func getMethodColor(method string) string {
	switch method {
	case fiber.MethodGet:
		return Blue
	case fiber.MethodPost:
		return Green
	case fiber.MethodPut:
		return Yellow
	case fiber.MethodDelete:
		return Red
	default:
		return Gray
	}
}

func CustomLogger(c *fiber.Ctx) error {
	start := time.Now()

	err := c.Next()

	stop := time.Now()
	latency := stop.Sub(start)

	status := c.Response().StatusCode()
	method := c.Method()
	path := c.OriginalURL()
	ip := c.IP()

	statusColor := getStatusColor(status)
	methodColor := getMethodColor(method)

	fmt.Printf(
		"%s[%s]%s %s%s%s %s - %s%d%s - %s - %s\n",
		Gray, stop.Format("15:04:05"), Reset,
		methodColor, method, Reset,
		path,
		statusColor, status, Reset,
		latency,
		ip,
	)

	return err
}
