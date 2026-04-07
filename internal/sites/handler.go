package sites

import (
	"github.com/gofiber/fiber/v2"
)

var service = SiteService{Repo: SiteRepository{}}

func ListSites(c *fiber.Ctx) error {
	s, err := service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(s)
}

func CreateSite(c *fiber.Ctx) error {
	type Req struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	s, err := service.Create(r.Name, r.Address)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(s)
}

func UpdateSite(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	type Req struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	s, err := service.Update(uuid, r.Name, r.Address)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(s)
}
