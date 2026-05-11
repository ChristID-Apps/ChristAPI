package contacts

import (
	"database/sql"
	"strconv"

	"christ-api/pkg/response"
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
			return response.Error(c, 422, "Invalid id", nil)
		}
		ct, err := service.GetByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return response.Error(c, 404, "Contact not found", nil)
			}
			return response.Error(c, 500, "Failed to retrieve contact", nil)
		}
		return response.Success(c, "Contact retrieved", ct)
	}

	out, err := service.List(page, limit)
	if err != nil {
		return response.Error(c, 500, "Failed to list contacts", nil)
	}
	meta := map[string]interface{}{"page": page, "limit": limit}
	return response.Paginated(c, "Contacts retrieved", out, meta)
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
		return response.Error(c, 422, "Invalid request", nil)
	}
	ct, err := service.Create(r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		return response.Error(c, 500, "Failed to create contact", nil)
	}
	return response.Created(c, "Contact created", ct)
}

func UpdateContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}
	type Req struct {
		FullName string  `json:"full_name"`
		Phone    *string `json:"phone"`
		Address  *string `json:"address"`
		SiteID   *int64  `json:"site_id"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.Error(c, 422, "Invalid request", nil)
	}
	ct, err := service.Update(id, r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.Error(c, 500, "Failed to update contact", nil)
	}
	return response.Success(c, "Contact updated", ct)
}

func DeleteContact(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid id", nil)
	}

	ct, err := service.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.Error(c, 500, "Failed to delete contact", nil)
	}
	return response.Success(c, "Contact deleted", ct)
}
