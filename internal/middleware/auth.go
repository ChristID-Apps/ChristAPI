package middleware

import (
	"strings"

	jwtpkg "christ-api/pkg/jwt"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

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

	token, err := jwtlib.Parse(tokenString[1], func(t *jwtlib.Token) (interface{}, error) {
		return jwtpkg.Secret(), nil
	})

	if err != nil || !token.Valid {
		return c.Status(401).JSON(fiber.Map{
			"error": "invalid token",
		})
	}

	// try to extract user_id claim and set to locals for handlers
	if claims, ok := token.Claims.(jwtlib.MapClaims); ok {
		if uid, exists := claims["user_id"]; exists {
			switch v := uid.(type) {
			case float64:
				c.Locals("user_id", int64(v))
			case int64:
				c.Locals("user_id", v)
			case int:
				c.Locals("user_id", int64(v))
			}
		}
	}

	return c.Next()
}
