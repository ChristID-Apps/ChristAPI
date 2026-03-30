package routes

import (
    "christ-api/internal/auth"
    "christ-api/internal/middleware"

    "github.com/gofiber/fiber/v2"
)

func Setup(app *fiber.App) {
    api := app.Group("/api")

    // public route
    api.Post("/login", auth.Login)

    // protected route
    protected := api.Group("/", middleware.AuthMiddleware)

    protected.Get("/profile", func(c *fiber.Ctx) error {
        return c.JSON(fiber.Map{
            "message": "you are logged in",
        })
    })
}