package middleware

import (
	"christ-api/internal/role"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

func AdminOnly(roleService *role.RoleService) fiber.Handler {
	return func(c *fiber.Ctx) error {
		roleID, ok := c.Locals("role_id").(int64)
		if !ok || roleID < 1 {
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
