package sites

import (
	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

var service = SiteService{Repo: SiteRepository{}}

func ListSites(c *fiber.Ctx) error {
	s, err := service.List()
	if err != nil {
		return response.Error(c, 500, "Failed to list sites", nil)
	}
	return response.Success(c, "Sites retrieved", s)
}

func CreateSite(c *fiber.Ctx) error {
	type Req struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	s, err := service.Create(r.Name, r.Address)
	if err != nil {
		return response.Error(c, 500, "Failed to create site", nil)
	}
	return response.Created(c, "Site created", s)
}

func UpdateSite(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	type Req struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	s, err := service.Update(uuid, r.Name, r.Address)
	if err != nil {
		return response.Error(c, 500, "Failed to update site", nil)
	}
	return response.Success(c, "Site updated", s)
}
