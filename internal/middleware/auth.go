package middleware

import (
	"strings"

	"github.com/gofiber/fiber/v2"
	"github.com/golang-jwt/jwt/v5"
)

var secret = []byte("secret-key")

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return c.Status(401).JSON(fiber.Map{
			"error": "missing token",
		})
	}

	// format: Bearer TOKEN
	tokenString := strings.Split(authHeader, " ")
	if len(tokenString) != 2 {
		return c.Status(401).JSON(fiber.Map{
			"error": "invalid token format",
		})
	}

	token, err := jwt.Parse(tokenString[1], func(t *jwt.Token) (interface{}, error) {
		return secret, nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	return c.Next()
}
