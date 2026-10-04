package routes

import (
	"net/http/httptest"
	"strings"
	"testing"

	"christ-api/pkg/database"

	"github.com/gofiber/fiber/v2"
)

type routeCase struct {
	method  string
	path    string
	secured bool
}

// Keep this inventory explicit so route removals, method changes, and missing
// auth boundaries are visible in a focused test failure.
var apiRoutes = []routeCase{
	{method: "GET", path: "/api/bible/versions"},
	{method: "GET", path: "/api/bible/books"},
	{method: "GET", path: "/api/bible/:version/search"},
	{method: "GET", path: "/api/bible/:version/:book/:chapter/:verse"},
	{method: "GET", path: "/api/bible/:version/:book/:chapter"},
	{method: "POST", path: "/api/login"},
	{method: "POST", path: "/api/register"},
	{method: "POST", path: "/api/verify-otp"},
	{method: "POST", path: "/api/resend-otp"},
	{method: "POST", path: "/api/auth/google"},
	{method: "POST", path: "/api/auth/google/username"},
	{method: "GET", path: "/api/profile", secured: true},
	{method: "PATCH", path: "/api/profile", secured: true},
	{method: "POST", path: "/api/profile/photo", secured: true},
	{method: "POST", path: "/api/logout", secured: true},
	{method: "GET", path: "/api/admin/approvals", secured: true},
	{method: "POST", path: "/api/admin/approvals/:id/approve", secured: true},
	{method: "POST", path: "/api/admin/approvals/:id/reject", secured: true},
	{method: "GET", path: "/api/admin/roles", secured: true},
	{method: "POST", path: "/api/admin/roles", secured: true},
	{method: "PATCH", path: "/api/admin/roles/:id", secured: true},
	{method: "GET", path: "/api/admin/activities", secured: true},
	{method: "GET", path: "/api/admin/activities/:uuid", secured: true},
	{method: "POST", path: "/api/admin/activities", secured: true},
	{method: "PATCH", path: "/api/admin/activities/:uuid", secured: true},
	{method: "DELETE", path: "/api/admin/activities/:uuid", secured: true},
	{method: "POST", path: "/api/admin/activities/:uuid/image", secured: true},
	{method: "PATCH", path: "/api/admin/activities/:uuid/bible-config", secured: true},
	{method: "GET", path: "/api/admin/activity-submissions", secured: true},
	{method: "POST", path: "/api/admin/activity-submissions/:uuid/approve", secured: true},
	{method: "POST", path: "/api/admin/activity-submissions/:uuid/reject", secured: true},
	{method: "GET", path: "/api/admin/bible-reading-submissions", secured: true},
	{method: "POST", path: "/api/admin/bible-reading-submissions/:uuid/approve", secured: true},
	{method: "POST", path: "/api/admin/bible-reading-submissions/:uuid/reject", secured: true},
	{method: "POST", path: "/api/admin/sites", secured: true},
	{method: "PATCH", path: "/api/admin/sites/:uuid", secured: true},
	{method: "POST", path: "/api/admin/contacts", secured: true},
	{method: "PATCH", path: "/api/admin/contacts/:id", secured: true},
	{method: "DELETE", path: "/api/admin/contacts/:id", secured: true},
	{method: "GET", path: "/api/admin/contacts", secured: true},
	{method: "GET", path: "/api/admin/contacts/:id", secured: true},
	{method: "GET", path: "/api/admin/points", secured: true},
	{method: "POST", path: "/api/admin/points/earn", secured: true},
	{method: "GET", path: "/api/admin/rewards", secured: true},
	{method: "POST", path: "/api/admin/rewards", secured: true},
	{method: "PATCH", path: "/api/admin/rewards/:uuid", secured: true},
	{method: "POST", path: "/api/admin/rewards/:uuid/image", secured: true},
	{method: "GET", path: "/api/admin/reward-redemptions", secured: true},
	{method: "POST", path: "/api/admin/reward-redemptions/:uuid/approve", secured: true},
	{method: "POST", path: "/api/admin/reward-redemptions/:uuid/reject", secured: true},
	{method: "POST", path: "/api/admin/reward-redemptions/:uuid/complete", secured: true},
	{method: "POST", path: "/api/admin/news", secured: true},
	{method: "PATCH", path: "/api/admin/news/:uuid", secured: true},
	{method: "DELETE", path: "/api/admin/news/:uuid", secured: true},
	{method: "GET", path: "/api/activity-categories", secured: true},
	{method: "GET", path: "/api/activities", secured: true},
	{method: "GET", path: "/api/activities/:uuid", secured: true},
	{method: "POST", path: "/api/activities/:uuid/submit", secured: true},
	{method: "GET", path: "/api/activity-submissions/me", secured: true},
	{method: "POST", path: "/api/bible-reading-submissions", secured: true},
	{method: "GET", path: "/api/bible-reading-submissions/me", secured: true},
	{method: "GET", path: "/api/sites", secured: true},
	{method: "GET", path: "/api/points", secured: true},
	{method: "POST", path: "/api/points/spend", secured: true},
	{method: "GET", path: "/api/rewards", secured: true},
	{method: "POST", path: "/api/rewards/:uuid/redeem", secured: true},
	{method: "GET", path: "/api/reward-redemptions/me", secured: true},
	{method: "POST", path: "/api/attendance/check-in", secured: true},
	{method: "GET", path: "/api/attendance/me", secured: true},
	{method: "GET", path: "/api/attendance/summary", secured: true},
	{method: "GET", path: "/api/admin/attendance", secured: true},
	{method: "GET", path: "/api/admin/attendance/summary", secured: true},
	{method: "GET", path: "/api/streaks", secured: true},
	{method: "POST", path: "/api/streaks/check-in", secured: true},
	{method: "GET", path: "/api/admin/news", secured: true},
	{method: "GET", path: "/api/news"},
	{method: "POST", path: "/api/admin/news/:uuid/image", secured: true},
}

func newRouteTestApp(t *testing.T) *fiber.App {
	t.Helper()
	previousDB := database.DB
	database.DB = nil
	t.Cleanup(func() { database.DB = previousDB })

	app := fiber.New()
	Setup(app)
	return app
}

func TestSetupRegistersEveryDocumentedAPIRoute(t *testing.T) {
	app := newRouteTestApp(t)
	registered := make(map[string]bool)
	for _, route := range app.GetRoutes() {
		registered[route.Method+" "+route.Path] = true
	}

	for _, want := range apiRoutes {
		key := want.method + " " + want.path
		if !registered[key] {
			t.Errorf("route not registered: %s", key)
		}
	}
}

func TestProtectedRoutesRequireAuthentication(t *testing.T) {
	app := newRouteTestApp(t)
	for _, route := range apiRoutes {
		if !route.secured {
			continue
		}

		path := route.path
		for _, parameter := range []string{":id", ":uuid"} {
			for contains(path, parameter) {
				path = replaceFirst(path, parameter, "1")
			}
		}
		for _, parameter := range []string{":version", ":book", ":chapter", ":verse"} {
			for contains(path, parameter) {
				path = replaceFirst(path, parameter, "1")
			}
		}

		t.Run(route.method+" "+route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, path, nil)
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.StatusCode != fiber.StatusUnauthorized {
				t.Fatalf("status = %d, want %d for unauthenticated request", response.StatusCode, fiber.StatusUnauthorized)
			}
		})
	}
}

func TestPublicAuthRoutesReachRequestValidation(t *testing.T) {
	app := newRouteTestApp(t)
	for _, route := range apiRoutes {
		if route.secured || route.method != "POST" || strings.HasPrefix(route.path, "/api/bible/") {
			continue
		}
		t.Run(route.path, func(t *testing.T) {
			request := httptest.NewRequest(route.method, route.path, strings.NewReader(`{}`))
			request.Header.Set(fiber.HeaderContentType, fiber.MIMEApplicationJSON)
			response, err := app.Test(request)
			if err != nil {
				t.Fatalf("request failed: %v", err)
			}
			if response.StatusCode != fiber.StatusUnprocessableEntity {
				t.Fatalf("status = %d, want %d for invalid public request", response.StatusCode, fiber.StatusUnprocessableEntity)
			}
		})
	}
}

func TestPublicNewsRouteDoesNotRequireAuthentication(t *testing.T) {
	app := newRouteTestApp(t)
	request := httptest.NewRequest("GET", "/api/news?status=draft", nil)
	response, err := app.Test(request)
	if err != nil {
		t.Fatalf("request failed: %v", err)
	}
	if response.StatusCode == fiber.StatusUnauthorized {
		t.Fatalf("public news route unexpectedly requires authentication: status=%d", response.StatusCode)
	}
	// The route test uses no database, so the handler is expected to fail only
	// after reaching repository access. The real handler still forces published.
	if response.StatusCode != fiber.StatusInternalServerError {
		t.Fatalf("status = %d, want %d when database is unavailable", response.StatusCode, fiber.StatusInternalServerError)
	}
}

func contains(value, part string) bool {
	return strings.Contains(value, part)
}

func replaceFirst(value, old, replacement string) string {
	return strings.Replace(value, old, replacement, 1)
}
