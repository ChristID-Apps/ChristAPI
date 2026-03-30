package main

import (
    "christ-api/pkg/database"
    "christ-api/routes"
	"christ-api/internal/middleware"

    "github.com/gofiber/fiber/v2"
)

func main() {
    database.Connect()

    app := fiber.New()
	app.Use(middleware.CustomLogger)
    routes.Setup(app)

    app.Listen(":3000")
}