package contacts

import (
	"database/sql"
	"encoding/json"
	"strconv"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *ContactService
}

func NewHandler(repo *ContactRepository) *Handler {
	return &Handler{service: &ContactService{Repo: repo}}
}

type ContactRequest struct {
	FullName string  `json:"full_name"`
	Phone    *string `json:"phone"`
	Address  *string `json:"address"`
	SiteID   *int64  `json:"site_id"`
}

type UpdateContactRequest struct {
	FullName *string         `json:"full_name"`
	Phone    *string         `json:"phone"`
	Address  *string         `json:"address"`
	SiteID   *int64          `json:"site_id"`
	Present  map[string]bool `json:"-"`
}

func (r *UpdateContactRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateContactRequest
	var decoded alias
	if err := json.Unmarshal(data, &decoded); err != nil {
		return err
	}
	var fields map[string]json.RawMessage
	if err := json.Unmarshal(data, &fields); err != nil {
		return err
	}
	decoded.Present = make(map[string]bool, len(fields))
	for field := range fields {
		decoded.Present[field] = true
	}
	*r = UpdateContactRequest(decoded)
	return nil
}

func (h *Handler) List(c *fiber.Ctx) error {
	page, _ := strconv.Atoi(c.Query("page", "1"))
	limit, _ := strconv.Atoi(c.Query("limit", "10"))

	idStr := c.Params("id")
	if idStr == "" {
		idStr = c.Query("id")
	}
	if idStr != "" {
		id, err := strconv.ParseInt(idStr, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid contact id", err)
		}
		ct, err := h.service.GetByID(id)
		if err != nil {
			if err == sql.ErrNoRows {
				return response.Error(c, 404, "Contact not found", nil)
			}
			return response.ErrorDetail(c, 500, "Failed to retrieve contact", err)
		}
		return response.Success(c, "Contact retrieved", ct)
	}

	out, err := h.service.List(page, limit)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list contacts", err)
	}
	meta := map[string]interface{}{"page": page, "limit": limit}
	return response.Paginated(c, "Contacts retrieved", out, meta)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	r := new(ContactRequest)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact request", err)
	}
	ct, err := h.service.Create(r.FullName, r.Phone, r.Address, r.SiteID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to create contact", err)
	}
	return response.Created(c, "Contact created", ct)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact id", err)
	}
	r := new(UpdateContactRequest)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact request", err)
	}
	ct, err := h.service.Update(id, r)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to update contact", err)
	}
	return response.Success(c, "Contact updated", ct)
}

func (h *Handler) Delete(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid contact id", err)
	}

	ct, err := h.service.Delete(id)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "Contact not found", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to delete contact", err)
	}
	return response.Success(c, "Contact deleted", ct)
}
