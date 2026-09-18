package points

import (
	"database/sql"
	"errors"
	"strconv"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

type Handler struct {
	service *Service
}

func NewHandler(repo *Repository) *Handler {
	return &Handler{service: &Service{Repo: repo}}
}

func parseLimit(raw string, defaultLimit int) (int, error) {
	if raw == "" {
		return defaultLimit, nil
	}
	v, err := strconv.Atoi(raw)
	if err != nil {
		return 0, err
	}
	switch v {
	case 5, 10, 20, 50:
		return v, nil
	default:
		return 0, fiber.NewError(400, "limit must be one of: 5, 10, 20, 50")
	}
}

func currentUserID(c *fiber.Ctx) (int64, bool) {
	v := c.Locals("user_id")
	switch id := v.(type) {
	case int64:
		return id, true
	case int:
		return int64(id), true
	case float64:
		return int64(id), true
	default:
		return 0, false
	}
}

func (h *Handler) Get(c *fiber.Ctx) error {
	currentID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}
	isAdmin, _ := c.Locals("is_admin").(bool)

	offset, err := parseOffset(c.Query("offset"))
	if err != nil {
		return response.Error(c, 422, err.Error(), nil)
	}

	var siteID *int64
	if siteRaw := c.Query("siteId"); siteRaw != "" {
		parsed, err := strconv.ParseInt(siteRaw, 10, 64)
		if err != nil {
			return response.ErrorDetail(c, 422, "Invalid siteId", err)
		}
		siteID = &parsed
	}

	userIDRaw := c.Query("userId")
	if userIDRaw == "" {
		if !isAdmin {
			userIDRaw = strconv.FormatInt(currentID, 10)
		} else {
			listLimit, err := parseLimit(c.Query("limit"), 10)
			if err != nil {
				return response.ErrorDetail(c, 422, "Invalid limit", err)
			}

			users, err := h.service.ListBalances(siteID, offset, listLimit)
			if err != nil {
				return response.ErrorDetail(c, 500, "Failed to list balances", err)
			}

			meta := map[string]interface{}{"offset": offset, "limit": listLimit}
			return response.Paginated(c, "Balances retrieved", users, meta)
		}
	}
	if !isAdmin {
		requestedID, err := strconv.ParseInt(userIDRaw, 10, 64)
		if err != nil || requestedID != currentID {
			return response.Error(c, 403, "You can only view your own points", nil)
		}
	}
	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid userId", err)
	}

	historyLimit, err := parseLimit(c.Query("limit"), 20)
	if err != nil {
		return response.ErrorDetail(c, 422, "Invalid limit", err)
	}

	state, err := h.service.GetState(userID, siteID, offset, historyLimit)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "User not found for given filters", nil)
		}
		return response.ErrorDetail(c, 500, "Failed to retrieve user state", err)
	}

	data := map[string]interface{}{"UserId": state.UserID, "Balance": state.Balance, "History": state.History}
	meta := map[string]interface{}{"offset": offset, "limit": historyLimit}
	return response.Paginated(c, "User points retrieved", data, meta)
}

type MutateRequest struct {
	Amount      int64   `json:"amount"`
	Reason      string  `json:"reason"`
	ReferenceID *string `json:"reference_id"`
}

func (h *Handler) Earn(c *fiber.Ctx) error {
	if isAdmin, _ := c.Locals("is_admin").(bool); !isAdmin {
		return response.Error(c, 403, "Admin authorization required", nil)
	}
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	var req MutateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid points request body", err)
	}

	entry, err := h.service.Earn(userID, req.Amount, req.Reason, req.ReferenceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, 404, "User not found", nil)
		}
		if errors.Is(err, ErrInsufficientPoints) {
			return response.Error(c, 409, err.Error(), nil)
		}
		if req.Amount <= 0 {
			return response.Error(c, 422, err.Error(), nil)
		}
		return response.ErrorDetail(c, 400, "Failed to earn points", err)
	}

	return response.Created(c, "Points earned", entry)
}

func (h *Handler) Spend(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	var req MutateRequest
	if err := c.BodyParser(&req); err != nil {
		return response.ErrorDetail(c, 422, "Invalid points request body", err)
	}

	entry, err := h.service.Spend(userID, req.Amount, req.Reason, req.ReferenceID)
	if err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return response.Error(c, 404, "User not found", nil)
		}
		if errors.Is(err, ErrInsufficientPoints) {
			return response.Error(c, 409, err.Error(), nil)
		}
		if req.Amount <= 0 {
			return response.Error(c, 422, err.Error(), nil)
		}
		return response.ErrorDetail(c, 400, "Failed to spend points", err)
	}

	return response.Created(c, "Points spent", entry)
}

func parseOffset(raw string) (int, error) {
	if raw == "" {
		return 0, nil
	}
	offset, err := strconv.Atoi(raw)
	if err != nil || offset < 0 {
		return 0, errors.New("offset must be a non-negative integer")
	}
	return offset, nil
}
