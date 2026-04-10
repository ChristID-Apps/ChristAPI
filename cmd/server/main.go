package main

import (
	"log"

	"christ-api/internal/auth"
	"christ-api/internal/middleware"
	"christ-api/pkg/database"
	"christ-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/joho/godotenv"
)

func main() {
	// load env dulu
	if err := godotenv.Load(); err != nil {
		log.Fatal("❌ .env tidak terbaca")
	}

	// connect database
	database.Connect()

	// initialize services that need DB
	auth.InitService(&auth.AuthRepository{DB: database.DB})

	app := fiber.New()
	app.Use(middleware.CustomLogger)

	routes.Setup(app)

	log.Println("🚀 Server running on :3000")
	app.Listen(":3000")
}
