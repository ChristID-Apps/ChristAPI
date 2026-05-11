package role

import (
	"strconv"

	"christ-api/pkg/response"
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
			return response.Error(c, 422, "Invalid id", nil)
		}
	}
	if siteStr := c.Query("siteId"); siteStr != "" {
		if v, err := strconv.ParseInt(siteStr, 10, 64); err == nil {
			sitePtr = &v
		} else {
			return response.Error(c, 422, "Invalid site_id", nil)
		}
	}

	roles, err := service.List(idPtr, sitePtr)
	if err != nil {
		return response.Error(c, 500, "Failed to list roles", nil)
	}
	return response.Success(c, "Roles retrieved", roles)
}

func CreateRole(c *fiber.Ctx) error {
	type Req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
		SiteID      *int64  `json:"site_id"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	rl, err := service.Create(r.Name, r.Description, r.SiteID)
	if err != nil {
		return response.Error(c, 500, "Failed to create role", nil)
	}
	return response.Created(c, "Role created", rl)
}

func UpdateRole(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	type Req struct {
		Name        string  `json:"name"`
		Description *string `json:"description"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	rl, err := service.Update(id, r.Name, r.Description)
	if err != nil {
		return response.Error(c, 500, "Failed to update role", nil)
	}
	return response.Success(c, "Role updated", rl)
}
