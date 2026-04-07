package routes

import (
	"christ-api/internal/auth"
	"christ-api/internal/contacts"
	"christ-api/internal/middleware"
	"christ-api/internal/role"
	"christ-api/internal/sites"

	"github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
	api := app.Group("/api")

	// public route
	api.Post("/login", auth.Login)
	api.Post("/register", auth.Register)

	// protected route
	protected := api.Group("/", middleware.AuthMiddleware)

	protected.Get("/profile", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"message": "you are logged in",
		})
	})

	// roles
	protected.Get("/roles", role.ListRoles)
	protected.Post("/roles", role.CreateRole)
	protected.Patch("/roles/:id", role.UpdateRole)

	// sites
	protected.Get("/sites", sites.ListSites)
	protected.Post("/sites", sites.CreateSite)
	protected.Patch("/sites/:uuid", sites.UpdateSite)

	// contacts
	protected.Get("/contacts", contacts.ListContacts)
	protected.Post("/contacts", contacts.CreateContact)
	protected.Patch("/contacts/:id", contacts.UpdateContact)
}
