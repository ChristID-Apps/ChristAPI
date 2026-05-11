package middleware

import (
	"strings"

	jwtpkg "christ-api/pkg/jwt"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func AuthMiddleware(c *fiber.Ctx) error {
	authHeader := c.Get("Authorization")

	if authHeader == "" {
		return response.Error(c, 401, "missing token", nil)
	}

	// format: Bearer TOKEN
	tokenString := strings.Split(authHeader, " ")
	if len(tokenString) != 2 {
		return response.Error(c, 401, "invalid token format", nil)
	}

	token, err := jwtlib.Parse(tokenString[1], func(t *jwtlib.Token) (interface{}, error) {
		return jwtpkg.Secret(), nil
	})

	if err != nil || !token.Valid {
		return response.Error(c, 401, "invalid token", nil)
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
