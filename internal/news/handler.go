package news

import (
	"strconv"

	"christ-api/pkg/response"
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
		return response.Error(c, 500, "Failed to list news", nil)
	}
	return response.Success(c, "News retrieved", out)
}

func CreateNews(c *fiber.Ctx) error {
	var req News
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	created, err := service.Create(&req)
	if err != nil {
		return response.Error(c, 500, "Failed to create news", nil)
	}
	return response.Created(c, "News created", created)
}

func UpdateNews(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	var req News
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	req.UUID = uuid
	if err := service.Update(&req); err != nil {
		return response.Error(c, 500, "Failed to update news", nil)
	}
	return response.Success(c, "News updated", nil)
}

func DeleteNews(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	if uuid == "" {
		return response.Error(c, 422, "uuid required", nil)
	}
	if err := service.Delete(uuid); err != nil {
		return response.Error(c, 500, "Failed to delete news", nil)
	}
	return response.Success(c, "News deleted", nil)
}
