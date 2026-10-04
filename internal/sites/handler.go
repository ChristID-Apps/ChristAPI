package sites

import (
	"christ-api/pkg/response"
	"encoding/json"
	"errors"
	"math"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *SiteService
}

type UpdateSiteRequest struct {
	Name      *string         `json:"name"`
	Address   *string         `json:"address"`
	Latitude  *float64        `json:"latitude"`
	Longitude *float64        `json:"longitude"`
	Present   map[string]bool `json:"-"`
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
		Name      string   `json:"name"`
		Address   *string  `json:"address"`
		Latitude  *float64 `json:"latitude"`
		Longitude *float64 `json:"longitude"`
	}
	r := new(Req)
	if err := c.BodyParser(r); err != nil {
		return response.ErrorDetail(c, 422, "Invalid site request", err)
	}
	if err := validateCoordinatePair(r.Latitude, r.Longitude, true, true); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}
	s, err := h.service.Create(r.Name, r.Address, r.Latitude, r.Longitude)
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
	if err := validateCoordinatePair(r.Latitude, r.Longitude, r.Present["latitude"], r.Present["longitude"]); err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}
	s, err := h.service.Update(uuid, r)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to update site", err)
	}
	return response.Success(c, "Site updated", s)
}

func validateCoordinatePair(latitude, longitude *float64, latitudePresent, longitudePresent bool) error {
	if !latitudePresent && !longitudePresent {
		return nil
	}
	if latitudePresent != longitudePresent {
		return errors.New("latitude and longitude must be provided together")
	}
	if latitude == nil && longitude == nil {
		return nil
	}
	if latitude == nil || longitude == nil || math.IsNaN(*latitude) || math.IsInf(*latitude, 0) ||
		math.IsNaN(*longitude) || math.IsInf(*longitude, 0) || *latitude < -90 || *latitude > 90 || *longitude < -180 || *longitude > 180 {
		return errors.New("latitude/longitude are outside valid coordinate ranges")
	}
	return nil
}
