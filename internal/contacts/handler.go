package contacts

import (
	"database/sql"
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var service = ContactService{Repo: ContactRepository{}}

func ListContacts(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	idStr := c.Params("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
		}
		ct, err := service.GetByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
			}
			return c.Status(500).JSON(fiber.Map{"error": err.Error()})
		}
		return c.JSON(ct)
	}

	out, err := service.List(page, limit)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(fiber.Map{
		"page":  page,
		"limit": limit,
		"data":  out,
	})
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
		SiteID   *int64  `json:"site_id"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	ct, err := service.Update(id, r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ct)
}

func DeleteContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid id"})
	}

	ct, err := service.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return c.Status(404).JSON(fiber.Map{"error": "contact not found"})
		}
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(ct)
}
