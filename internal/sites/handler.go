package sites

import (
	"christ-api/pkg/response"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *SiteService
}

type UpdateSiteRequest struct {
	Name    *string         `json:"name"`
	Address *string         `json:"address"`
	Present map[string]bool `json:"-"`
}

func (r *UpdateSiteRequest) UnmarshalJSON(data []byte) error {
	type alias UpdateSiteRequest
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
	*r = UpdateSiteRequest(decoded)
	return nil
}

func NewHandler(repo *SiteRepository) *Handler {
	return &Handler{service: &SiteService{Repo: repo}}
}

func (h *Handler) List(c *fiber.Ctx) error {
	s, err := h.service.List()
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list sites", err)
	}
	return response.Success(c, "Sites retrieved", s)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	type Req struct {
		Name    string  `json:"name"`
		Address *string `json:"address"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid site request", err)
	}
	s, err := h.service.Create(r.Name, r.Address)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to create site", err)
	}
	return response.Created(c, "Site created", s)
}

func (h *Handler) Update(c *fiber.Ctx) error {
	uuid := c.Params("uuid")
	r := new(UpdateSiteRequest)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid site request", err)
	}
	s, err := h.service.Update(uuid, r)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to update site", err)
	}
	return response.Success(c, "Site updated", s)
}
