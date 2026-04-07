package contacts

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var service = ContactService{Repo: ContactRepository{}}

func ListContacts(c *fiber.Ctx) error {
	out, err := service.List()
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func CreateContact(c *fiber.Ctx) error {
	type Req struct {
		FullName string  `json:"full_name"`
		Phone    *string `json:"phone"`
		Address  *string `json:"address"`
		SiteID   *int64  `json:"site_id"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	ct, err := service.Create(r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(ct)
}

func UpdateContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}
	type Req struct {
		FullName string  `json:"full_name"`
		Phone    *string `json:"phone"`
		Address  *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	ct, err := service.Update(id, r.FullName, r.Phone, r.Address)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ct)
}
