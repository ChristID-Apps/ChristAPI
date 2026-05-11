package points

import (
	"database/sql"
	"strconv"

	"christ-api/pkg/response"
	"github.com/gofiber/fiber/v2"
)

var service = Service{Repo: Repository{}}

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

func GetPoints(c *fiber.Ctx) error {
	_, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	offset, _ := strconv.Atoi(c.Query("offset", "0"))
	if offset < 0 {
		offset = 0
	}

	var siteID *int64
	if siteRaw := c.Query("siteId"); siteRaw != "" {
		parsed, err := strconv.ParseInt(siteRaw, 10, 64)
		if err != nil {
			return response.Error(c, 422, "Invalid siteId", nil)
		}
		siteID = &parsed
	}

	userIDRaw := c.Query("userId")
	if userIDRaw == "" {
		listLimit, err := parseLimit(c.Query("limit"), 10)
		if err != nil {
			return response.Error(c, 422, "Invalid limit", nil)
		}

		users, err := service.ListBalances(siteID, offset, listLimit)
		if err != nil {
			return response.Error(c, 500, "Failed to list balances", nil)
		}

		meta := map[string]interface{}{"offset": offset, "limit": listLimit}
		return response.Paginated(c, "Balances retrieved", users, meta)
	}

	userID, err := strconv.ParseInt(userIDRaw, 10, 64)
	if err != nil {
		return response.Error(c, 422, "Invalid userId", nil)
	}

	historyLimit, err := parseLimit(c.Query("limit"), 20)
	if err != nil {
		return response.Error(c, 422, "Invalid limit", nil)
	}

	state, err := service.GetState(userID, siteID, offset, historyLimit)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "User not found for given filters", nil)
		}
		return response.Error(c, 500, "Failed to retrieve user state", nil)
	}

	data := map[string]interface{}{"UserId": state.UserID, "Balance": state.Balance, "History": state.History}
	meta := map[string]interface{}{"offset": offset, "limit": historyLimit}
	return response.Paginated(c, "User points retrieved", data, meta)
}

type mutatePointsRequest struct {
	Amount      int64   `json:"amount"`
	Reason      string  `json:"reason"`
	ReferenceID *string `json:"reference_id"`
}

func EarnPoints(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	var req mutatePointsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, 422, "Invalid request body", nil)
	}

	entry, err := service.Earn(userID, req.Amount, req.Reason, req.ReferenceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "User not found", nil)
		}
		return response.Error(c, 400, "Failed to earn points", nil)
	}

	return response.Created(c, "Points earned", entry)
}

func SpendPoints(c *fiber.Ctx) error {
	userID, ok := currentUserID(c)
	if !ok {
		return response.Error(c, 401, "Unauthorized", nil)
	}

	var req mutatePointsRequest
	if err := c.BodyParser(&req); err != nil {
		return response.Error(c, 422, "Invalid request body", nil)
	}

	entry, err := service.Spend(userID, req.Amount, req.Reason, req.ReferenceID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Error(c, 404, "User not found", nil)
		}
		return response.Error(c, 400, "Failed to spend points", nil)
	}

	return response.Created(c, "Points spent", entry)
}
