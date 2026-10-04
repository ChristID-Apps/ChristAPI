package response

import (
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gofiber/fiber/v2"
)

func TestErrorDetailDoesNotExposeInternalError(t *testing.T) {
	app := fiber.New()
	app.Get("/", func(c *fiber.Ctx) error {
		return ErrorDetail(c, http.StatusInternalServerError, "Request failed", errors.New("database password leaked"))
	})

	request := httptest.NewRequest(http.MethodGet, "/", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatal(err)
	}
	defer response.Body.Close()

	body, err := io.ReadAll(response.Body)
	if err != nil {
		t.Fatal(err)
	}
	if strings.Contains(string(body), "database password leaked") {
		t.Fatal("response exposed the internal error")
	}
	if !strings.Contains(string(body), "Internal server error") {
		t.Fatalf("response does not contain the safe error detail: %s", body)
	}
}
