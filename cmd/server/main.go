package main

import (
	"log"
	"os"

	"christ-api/internal/middleware"
	"christ-api/pkg/database"
	"christ-api/routes"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// load env
	if err := godotenv.Load(".env.local"); err != nil {
		if err := godotenv.Load(".env"); err != nil {
			log.Println("ℹ️ .env.local/.env tidak ditemukan, pakai environment variables")
		}
	}

	if os.Getenv("JWT_SECRET") == "" {
		log.Fatal("JWT_SECRET wajib diisi")
	}

	// connect database
	database.Connect()

	app := fiber.New()

	// CORS with restricted origins
	corsOrigins := os.Getenv("CORS_ORIGINS")
	if corsOrigins == "" {
		corsOrigins = "http://localhost:3000,http://localhost:3001"
	}
	app.Use(cors.New(cors.Config{
		AllowOrigins: corsOrigins,
	}))

	app.Use(middleware.CustomLogger)
	app.Use(middleware.RateLimiter)

	// Serve static files from docs directory
	app.Static("/docs", "./docs")
	app.Static("/uploads", "./uploads")

	routes.Setup(app)

	port := os.Getenv("API_PORT")
	if port == "" {
		port = "3000"
	}

	log.Println("🚀 Server running on :" + port)
	if err := app.Listen(":" + port); err != nil {
		log.Fatal("❌ Failed to start server:", err)
	}
}
