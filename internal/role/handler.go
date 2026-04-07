package role

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var service = RoleService{Repo: RoleRepository{}}

func ListRoles(c *fiber.Ctx) error {
	var idPtr *int64
	var sitePtr *int64

	if idStr := c.Query("id"); idStr != "" {
		if v, err := strconv.ParseInt(idStr, 10, 64); err == nil {
			idPtr = &v
		} else {
			return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
		}
	}
	if siteStr := c.Query("siteId"); siteStr != "" {
		if v, err := strconv.ParseInt(siteStr, 10, 64); err == nil {
			sitePtr = &v
		} else {
			return c.Status(400).JSON(fiber.Map{"error": "invalid site_id"})
		}
	}

	roles, err := service.List(idPtr, sitePtr)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(roles)
}

func CreateRole(c *fiber.Ctx) error {
	type Req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		SiteID      *int64  `json:"site_id"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	rl, err := service.Create(r.Name, r.Description, r.SiteID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(rl)
}

func UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	type Req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	rl, err := service.Update(id, r.Name, r.Description)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(rl)
}
