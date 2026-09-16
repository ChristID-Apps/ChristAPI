package middleware

import (
	"strings"

	"christ-api/internal/role"
	jwtpkg "christ-api/pkg/jwt"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
	jwtlib "github.com/golang-jwt/jwt/v5"
)

func AdminOnly(roleService *role.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		authHeader := c.Get("Authorization")
		if authHeader == "" {
			return response.Error(c, 401, "missing token", nil)
		}

		tokenString := strings.Split(authHeader, " ")
		if len(tokenString) != 2 {
			return response.Error(c, 401, "invalid token format", fiber.Map{"detail": "Authorization must use the Bearer <token> format"})
		}

		token, err := jwtlib.Parse(tokenString[1], func(t *jwtlib.Token) (interface{}, error) {
			return jwtpkg.Secret(), nil
		}, jwtlib.WithValidMethods([]string{jwtlib.SigningMethodHS256.Alg()}))

		if err != nil || !token.Valid {
			return response.ErrorDetail(c, 401, "invalid token", err)
		}

		var roleID int64
		if claims, ok := token.Claims.(jwtlib.MapClaims); ok {
			if rid, exists := claims["role_id"]; exists {
				switch v := rid.(type) {
				case float64:
					roleID = int64(v)
				case int64:
					roleID = v
				}
			}
		}

		if roleID == 0 {
			return response.Error(c, 403, "no role assigned", fiber.Map{"detail": "role_id claim is missing or zero"})
		}

		roleObj, err := roleService.GetByID(roleID)
		if err != nil {
			return response.ErrorDetail(c, 403, "admin role lookup failed", err)
		}

		if roleObj.Code != "admin" && roleObj.Code != "super_admin" {
			return response.Error(c, 403, "admin role required", fiber.Map{"detail": "current role code is not admin or super_admin", "role_code": roleObj.Code})
		}

		c.Locals("is_admin", true)
		return c.Next()
	}
}
