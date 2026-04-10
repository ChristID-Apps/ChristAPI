package news

import (
	"strconv"

	"github.com/gofiber/fiber/v2"
)

var service = NewsService{Repo: NewsRepository{}}

func ListNews(c *fiber.Ctx) error {
	var filter NewsFilter

	if v := c.Query("site_id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.SiteID = &id
		}
	}
	if v := c.Query("id"); v != "" {
		if id, err := strconv.ParseInt(v, 10, 64); err == nil {
			filter.ID = &id
		}
	}
	if v := c.Query("search"); v != "" {
		filter.Search = &v
	}
	if v := c.Query("limit"); v != "" {
		if lim, err := strconv.Atoi(v); err == nil {
			filter.Limit = lim
		}
	}
	if v := c.Query("offset"); v != "" {
		if off, err := strconv.Atoi(v); err == nil {
			filter.Offset = off
		}
	}

	out, err := service.List(filter)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.JSON(out)
}

func CreateNews(c *fiber.Ctx) error {
	var req News
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	created, err := service.Create(&req)
	if err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.Status(201).JSON(created)
}

func UpdateNews(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	var req News
	if err := c.BodyParser(&req); err != nil {
		return c.Status(400).JSON(fiber.Map{"error": "invalid request"})
	}
	req.UUID = uuid
	if err := service.Update(&req); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}

func DeleteNews(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return c.Status(400).JSON(fiber.Map{"error": "uuid required"})
	}
	if err := service.Delete(uuid); err != nil {
		return c.Status(500).JSON(fiber.Map{"error": err.Error()})
	}
	return c.SendStatus(204)
}
