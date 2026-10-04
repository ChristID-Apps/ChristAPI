package attendance

import (
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TestParseDateOrEmpty(t *testing.T) {
	tests := []struct {
		name      string
		input     string
		want      string
		wantError bool
	}{
		{name: "empty date is omitted"},
		{name: "valid date", input: "2026-10-01", want: "2026-10-01"},
		{name: "invalid date", input: "2026-02-30", wantError: true},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			got, err := parseDateOrEmpty(test.input)
			if (err != nil) != test.wantError {
				t.Fatalf("error = %v, wantError = %t", err, test.wantError)
			}
			if got != test.want {
				t.Fatalf("date = %q, want %q", got, test.want)
			}
		})
	}
}

func TestBusinessDateUsesAsiaJakarta(t *testing.T) {
	input := time.Date(2026, time.October, 5, 18, 30, 0, 0, time.UTC)
	got, err := businessDate(input)
	if err != nil {
		t.Fatal(err)
	}
	if got != "2026-10-06" {
		t.Fatalf("business date = %q, want 2026-10-06", got)
	}
}

func TestCheckInRequiresDeviceLocation(t *testing.T) {
	app := fiber.New()
	handler := NewHandler(&Repository{})
	app.Post("/check-in", func(c *fiber.Ctx) error {
		c.Locals("user_id", int64(9))
		return handler.CheckIn(c)
	})
	request := httptest.NewRequest("POST", "/check-in", strings.NewReader(`{"site_id":1}`))
	request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	if response.StatusCode != fiber.StatusUnprocessableEntity {
		t.Fatalf("status = %d, want %d", response.StatusCode, fiber.StatusUnprocessableEntity)
	}
}
