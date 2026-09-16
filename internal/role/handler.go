package role

import (
	"strconv"

	"christ-api/internal/role/dto/requests"
	"christ-api/internal/role/helpers"
	"christ-api/pkg/response"

	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *RoleService
}

func NewHandler(repo *RoleRepository) *Handler {
	return &Handler{service: &RoleService{Repo: repo}}
}

func (h *Handler) List(c *fiber.Ctx) error {
	req := &requests.ListRolesRequest{}

	// check request query parameters
	if err := c.QueryParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid role query parameters", err)
	}

	// validasiin request query parameters
	roles, err := h.service.List(req.ID, req.SiteID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to list roles", err)
	}

	// convert roles to response DTOs
	resp := RolesToRoleResponses(roles)
	return response.Success(c, "Roles retrieved", resp)
}

func (h *Handler) Create(c *fiber.Ctx) error {
	req := new(requests.CreateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid role request", err)
	}

	if err := helpers.ValidateCreateRoleRequest(req); err != nil {
		return response.ErrorDetail(c, 422, "Role validation failed", err)
	}

	role, err := h.service.Create(req.Name, req.Code, req.Description, req.SiteID)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to create role", err)
	}

	return response.Created(c, "Role created", RoleToRoleResponse(role))
}

func (h *Handler) Update(c *fiber.Ctx) error {
	idStr := c.Params("id")
	id, err := strconv.ParseInt(idStr, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid role ID", err)
	}

	req := new(requests.UpdateRoleRequest)
	if err := c.BodyParser(req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid role request", err)
	}

	if err := helpers.ValidateUpdateRoleRequest(req); err != nil {
		return response.ErrorDetail(c, 422, "Role validation failed", err)
	}

	role, err := h.service.Update(id, req)
	if err != nil {
		return response.ErrorDetail(c, 500, "Failed to update role", err)
	}

	return response.Success(c, "Role updated", RoleToRoleResponse(role))
}
